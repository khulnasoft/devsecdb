package v1

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/khulnasoft/devsecdb/backend/api/auth"
	"github.com/khulnasoft/devsecdb/backend/common"
	"github.com/khulnasoft/devsecdb/backend/common/log"
	"github.com/khulnasoft/devsecdb/backend/component/config"
	"github.com/khulnasoft/devsecdb/backend/component/iam"
	"github.com/khulnasoft/devsecdb/backend/component/state"
	enterprise "github.com/khulnasoft/devsecdb/backend/enterprise/api"
	api "github.com/khulnasoft/devsecdb/backend/legacyapi"
	metricapi "github.com/khulnasoft/devsecdb/backend/metric"
	"github.com/khulnasoft/devsecdb/backend/plugin/idp/ldap"
	"github.com/khulnasoft/devsecdb/backend/plugin/idp/oauth2"
	"github.com/khulnasoft/devsecdb/backend/plugin/idp/oidc"
	"github.com/khulnasoft/devsecdb/backend/plugin/metric"
	"github.com/khulnasoft/devsecdb/backend/runner/metricreport"
	"github.com/khulnasoft/devsecdb/backend/store"
	"github.com/khulnasoft/devsecdb/backend/utils"
	storepb "github.com/khulnasoft/devsecdb/proto/generated-go/store"
	v1pb "github.com/khulnasoft/devsecdb/proto/generated-go/v1"
)

type CreateUserFunc func(ctx context.Context, user *store.UserMessage, firstEndUser bool) error

var (
	invalidUserOrPasswordError = status.Errorf(codes.Unauthenticated, "The email or password is not valid.")
)

// AuthService implements the auth service.
type AuthService struct {
	v1pb.UnimplementedAuthServiceServer
	store          *store.Store
	secret         string
	licenseService enterprise.LicenseService
	metricReporter *metricreport.Reporter
	profile        *config.Profile
	stateCfg       *state.State
	iamManager     *iam.Manager
	postCreateUser func(ctx context.Context, user *store.UserMessage, firstEndUser bool) error
}

// NewAuthService creates a new AuthService.
func NewAuthService(store *store.Store, secret string, licenseService enterprise.LicenseService, metricReporter *metricreport.Reporter, profile *config.Profile, stateCfg *state.State, iamManager *iam.Manager, postCreateUser func(ctx context.Context, user *store.UserMessage, firstEndUser bool) error) (*AuthService, error) {
	return &AuthService{
		store:          store,
		secret:         secret,
		licenseService: licenseService,
		metricReporter: metricReporter,
		profile:        profile,
		stateCfg:       stateCfg,
		iamManager:     iamManager,
		postCreateUser: postCreateUser,
	}, nil
}

// GetUser gets a user.
func (s *AuthService) GetUser(ctx context.Context, request *v1pb.GetUserRequest) (*v1pb.User, error) {
	userID, err := common.GetUserID(request.Name)
	var user *store.UserMessage
	if err != nil {
		email, err := common.GetUserEmail(request.Name)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		u, err := s.store.GetUserByEmail(ctx, email)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get user, error: %v", err)
		}
		user = u
	} else {
		u, err := s.store.GetUserByID(ctx, userID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get user, error: %v", err)
		}
		user = u
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user %d not found", userID)
	}
	return convertToUser(user), nil
}

// ListUsers lists all users.
func (s *AuthService) ListUsers(ctx context.Context, request *v1pb.ListUsersRequest) (*v1pb.ListUsersResponse, error) {
	users, err := s.store.ListUsers(ctx, &store.FindUserMessage{ShowDeleted: request.ShowDeleted})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list user, error: %v", err)
	}
	response := &v1pb.ListUsersResponse{}
	for _, user := range users {
		response.Users = append(response.Users, convertToUser(user))
	}
	return response, nil
}

// CreateUser creates a user.
func (s *AuthService) CreateUser(ctx context.Context, request *v1pb.CreateUserRequest) (*v1pb.User, error) {
	if err := s.userCountGuard(ctx); err != nil {
		return nil, err
	}

	setting, err := s.store.GetWorkspaceGeneralSetting(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find workspace setting, error: %v", err)
	}

	if setting.DisallowSignup {
		callerUser, ok := ctx.Value(common.UserContextKey).(*store.UserMessage)
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "sign up is disallowed")
		}
		ok, err := s.iamManager.CheckPermission(ctx, iam.PermissionUsersCreate, callerUser)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "user does not have permission %q", iam.PermissionUsersCreate)
		}
	}
	if request.User == nil {
		return nil, status.Errorf(codes.InvalidArgument, "user must be set")
	}
	if request.User.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email must be set")
	}
	if request.User.Title == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user title must be set")
	}

	principalType, err := convertToPrincipalType(request.User.UserType)
	if err != nil {
		return nil, err
	}
	if request.User.UserType != v1pb.UserType_SERVICE_ACCOUNT && request.User.UserType != v1pb.UserType_USER {
		return nil, status.Errorf(codes.InvalidArgument, "support user and service account only")
	}

	count, err := s.store.CountUsers(ctx, api.EndUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count users, error: %v", err)
	}
	firstEndUser := count == 0

	if request.User.Phone != "" {
		if err := common.ValidatePhone(request.User.Phone); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid phone %q, error: %v", request.User.Phone, err)
		}
	}

	var allowedDomains []string
	if setting.EnforceIdentityDomain {
		allowedDomains = setting.Domains
	}
	if err := validateEmail(request.User.Email, allowedDomains, principalType == api.ServiceAccount); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email %q, error: %v", request.User.Email, err)
	}
	existingUser, err := s.store.GetUserByEmail(ctx, request.User.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find user by email, error: %v", err)
	}
	if existingUser != nil {
		return nil, status.Errorf(codes.AlreadyExists, "email %s is already existed", request.User.Email)
	}

	password := request.User.Password
	if request.User.UserType == v1pb.UserType_SERVICE_ACCOUNT {
		pwd, err := common.RandomString(20)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate access key for service account.")
		}
		password = fmt.Sprintf("%s%s", api.ServiceAccountAccessKeyPrefix, pwd)
	} else {
		if password != "" {
			if err := s.validatePassword(ctx, password); err != nil {
				return nil, err
			}
		} else {
			pwd, err := common.RandomString(20)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to generate random password for service account.")
			}
			password = pwd
		}
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate password hash, error: %v", err)
	}
	userMessage := &store.UserMessage{
		Email:        request.User.Email,
		Name:         request.User.Title,
		Phone:        request.User.Phone,
		Type:         principalType,
		PasswordHash: string(passwordHash),
	}

	creatorUID := api.SystemBotID
	u, ok := ctx.Value(common.UserContextKey).(*store.UserMessage)
	if ok && u != nil {
		creatorUID = u.ID
	}

	user, err := s.store.CreateUser(ctx, userMessage, creatorUID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user, error: %v", err)
	}

	if firstEndUser {
		// The first end user should be workspace admin.
		updateRole := &store.PatchIamPolicyMessage{
			Member:     common.FormatUserUID(user.ID),
			UpdaterUID: creatorUID,
			Roles:      []string{common.FormatRole(api.WorkspaceAdmin.String())},
		}
		if _, err := s.store.PatchWorkspaceIamPolicy(ctx, updateRole); err != nil {
			return nil, err
		}
	}

	if err := s.postCreateUser(ctx, user, firstEndUser); err != nil {
		return nil, err
	}

	isFirstUser := user.ID == api.PrincipalIDForFirstUser
	s.metricReporter.Report(ctx, &metric.Metric{
		Name:  metricapi.PrincipalRegistrationMetricName,
		Value: 1,
		Labels: map[string]any{
			"email": user.Email,
			"name":  user.Name,
			"phone": user.Phone,
			// We only send lark notification for the first principal registration.
			// false means do not notify upfront. Later the notification will be triggered by the scheduler.
			"lark_notified": !isFirstUser,
		},
	})

	userResponse := convertToUser(user)
	if request.User.UserType == v1pb.UserType_SERVICE_ACCOUNT {
		userResponse.ServiceKey = password
	}
	return userResponse, nil
}

func (s *AuthService) validatePassword(ctx context.Context, password string) error {
	passwordRestriction, err := s.store.GetPasswordRestrictionSetting(ctx)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get password restriction with error: %v", err)
	}
	if len(password) < int(passwordRestriction.MinLength) {
		return status.Errorf(codes.InvalidArgument, "password length should no less than %v characters", passwordRestriction.MinLength)
	}
	if passwordRestriction.RequireNumber && !regexp.MustCompile("[0-9]+").MatchString(password) {
		return status.Errorf(codes.InvalidArgument, "password must contains at least 1 number")
	}
	if passwordRestriction.RequireLetter && !regexp.MustCompile("[a-zA-Z]+").MatchString(password) {
		return status.Errorf(codes.InvalidArgument, "password must contains at least 1 lower case letter")
	}
	if passwordRestriction.RequireUppercaseLetter && !regexp.MustCompile("[A-Z]+").MatchString(password) {
		return status.Errorf(codes.InvalidArgument, "password must contains at least 1 upper case letter")
	}
	if passwordRestriction.RequireSpecialCharacter && !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]+`).MatchString(password) {
		return status.Errorf(codes.InvalidArgument, "password must contains at least 1 special character")
	}
	return nil
}

// UpdateUser updates a user.
func (s *AuthService) UpdateUser(ctx context.Context, request *v1pb.UpdateUserRequest) (*v1pb.User, error) {
	callerUser, ok := ctx.Value(common.UserContextKey).(*store.UserMessage)
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "failed to get caller user")
	}
	if request.User == nil {
		return nil, status.Errorf(codes.InvalidArgument, "user must be set")
	}
	if request.UpdateMask == nil {
		return nil, status.Errorf(codes.InvalidArgument, "update_mask must be set")
	}

	userID, err := common.GetUserID(request.User.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user, error: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user %d not found", userID)
	}
	if user.MemberDeleted {
		return nil, status.Errorf(codes.NotFound, "user %q has been deleted", userID)
	}

	if callerUser.ID != userID {
		ok, err := s.iamManager.CheckPermission(ctx, iam.PermissionUsersUpdate, callerUser)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "user does not have permission %q", iam.PermissionUsersUpdate)
		}
	}

	setting, err := s.store.GetWorkspaceGeneralSetting(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find workspace setting, error: %v", err)
	}

	var passwordPatch *string
	patch := &store.UpdateUserMessage{}
	for _, path := range request.UpdateMask.Paths {
		switch path {
		case "email":
			if user.Profile.Source != "" {
				return nil, status.Errorf(codes.InvalidArgument, "cannot change email for external user")
			}
			var allowedDomains []string
			if setting.EnforceIdentityDomain {
				allowedDomains = setting.Domains
			}
			if err := validateEmail(request.User.Email, allowedDomains, user.Type == api.ServiceAccount); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid email %q, error: %v", request.User.Email, err)
			}
			existedUser, err := s.store.GetUserByEmail(ctx, request.User.Email)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to find user list, error: %v", err)
			}
			if existedUser != nil && existedUser.ID != user.ID {
				return nil, status.Errorf(codes.AlreadyExists, "email %s is already existed", request.User.Email)
			}
			patch.Email = &request.User.Email
		case "title":
			patch.Name = &request.User.Title
		case "password":
			if user.Type != api.EndUser {
				return nil, status.Errorf(codes.InvalidArgument, "password can be mutated for end users only")
			}
			if err := s.validatePassword(ctx, request.User.Password); err != nil {
				return nil, err
			}
			passwordPatch = &request.User.Password
		case "service_key":
			if user.Type != api.ServiceAccount {
				return nil, status.Errorf(codes.InvalidArgument, "service key can be mutated for service accounts only")
			}
			val, err := common.RandomString(20)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to generate access key for service account.")
			}
			password := fmt.Sprintf("%s%s", api.ServiceAccountAccessKeyPrefix, val)
			passwordPatch = &password
		case "mfa_enabled":
			if request.User.MfaEnabled {
				if user.MFAConfig.TempOtpSecret == "" || len(user.MFAConfig.TempRecoveryCodes) == 0 {
					return nil, status.Errorf(codes.InvalidArgument, "MFA is not setup yet")
				}
				patch.MFAConfig = &storepb.MFAConfig{
					OtpSecret:     user.MFAConfig.TempOtpSecret,
					RecoveryCodes: user.MFAConfig.TempRecoveryCodes,
				}
			} else {
				setting, err := s.store.GetWorkspaceGeneralSetting(ctx)
				if err != nil {
					return nil, status.Errorf(codes.Internal, "failed to find workspace setting, error: %v", err)
				}
				if setting.Require_2Fa {
					isWorkspaceAdmin, err := s.isUserWorkspaceAdmin(ctx, callerUser)
					if err != nil {
						return nil, status.Errorf(codes.Internal, "failed to check user roles, error: %v", err)
					}
					// Allow workspace admin to disable 2FA even if it is required.
					if !isWorkspaceAdmin {
						return nil, status.Errorf(codes.InvalidArgument, "2FA is required and cannot be disabled")
					}
				}
				patch.MFAConfig = &storepb.MFAConfig{}
			}
		case "phone":
			if request.User.Phone != "" {
				if err := common.ValidatePhone(request.User.Phone); err != nil {
					return nil, status.Errorf(codes.InvalidArgument, "invalid phone number %q, error: %v", request.User.Phone, err)
				}
			}
			patch.Phone = &request.User.Phone
		}
	}
	if passwordPatch != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(*passwordPatch)); err == nil {
			// return bad request if the passwords match
			return nil, status.Errorf(codes.InvalidArgument, "password cannot be the same")
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte((*passwordPatch)), bcrypt.DefaultCost)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate password hash, error: %v", err)
		}
		passwordHashStr := string(passwordHash)
		patch.PasswordHash = &passwordHashStr
	}
	// This flag is mainly used for validating OTP code when user setup MFA.
	// We only validate OTP code but not update user.
	if request.OtpCode != nil {
		isValid := validateWithCodeAndSecret(*request.OtpCode, user.MFAConfig.TempOtpSecret)
		if !isValid {
			return nil, status.Errorf(codes.InvalidArgument, "invalid OTP code")
		}
	}
	// This flag will regenerate temp secret and temp recovery codes.
	// It will be used when user setup MFA and regenerating recovery codes.
	if request.RegenerateTempMfaSecret {
		tempSecret, err := generateRandSecret(user.Email)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate MFA secret, error: %v", err)
		}
		tempRecoveryCodes, err := generateRecoveryCodes(10)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate recovery codes, error: %v", err)
		}
		patch.MFAConfig = &storepb.MFAConfig{
			TempOtpSecret:     tempSecret,
			TempRecoveryCodes: tempRecoveryCodes,
		}
		if user.MFAConfig != nil {
			patch.MFAConfig.OtpSecret = user.MFAConfig.OtpSecret
			patch.MFAConfig.RecoveryCodes = user.MFAConfig.RecoveryCodes
		}
	}
	// This flag will update user's recovery codes with temp recovery codes.
	// It will be used when user regenerate recovery codes after two phase commit.
	if request.RegenerateRecoveryCodes {
		if user.MFAConfig.OtpSecret == "" {
			return nil, status.Errorf(codes.InvalidArgument, "MFA is not enabled")
		}
		if len(user.MFAConfig.TempRecoveryCodes) == 0 {
			return nil, status.Errorf(codes.InvalidArgument, "No recovery codes to update")
		}
		patch.MFAConfig = &storepb.MFAConfig{
			OtpSecret:     user.MFAConfig.OtpSecret,
			RecoveryCodes: user.MFAConfig.TempRecoveryCodes,
		}
	}

	user, err = s.store.UpdateUser(ctx, user, patch, callerUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user, error: %v", err)
	}

	userResponse := convertToUser(user)
	if request.User.UserType == v1pb.UserType_SERVICE_ACCOUNT && passwordPatch != nil {
		userResponse.ServiceKey = *passwordPatch
	}
	return userResponse, nil
}

// DeleteUser deletes a user.
func (s *AuthService) DeleteUser(ctx context.Context, request *v1pb.DeleteUserRequest) (*emptypb.Empty, error) {
	callerUser, ok := ctx.Value(common.UserContextKey).(*store.UserMessage)
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "failed to get caller user")
	}
	ok, err := s.iamManager.CheckPermission(ctx, iam.PermissionUsersDelete, callerUser)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "user does not have permission %q", iam.PermissionUsersDelete)
	}

	userID, err := common.GetUserID(request.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user, error: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user %d not found", userID)
	}
	if user.MemberDeleted {
		return nil, status.Errorf(codes.NotFound, "user %q has been deleted", userID)
	}

	// Check if there is still workspace admin if the current user is deleted.
	policy, err := s.store.GetWorkspaceIamPolicy(ctx)
	if err != nil {
		return nil, err
	}
	ok = hasExtraWorkspaceAdmin(policy.Policy, user.ID)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "workspace must have at least one admin")
	}

	if _, err := s.store.UpdateUser(ctx, user, &store.UpdateUserMessage{Delete: &deletePatch}, callerUser.ID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func hasExtraWorkspaceAdmin(policy *storepb.IamPolicy, userID int) bool {
	workspaceAdminRole := common.FormatRole(api.WorkspaceAdmin.String())
	userMember := common.FormatUserUID(userID)
	systemBotMember := common.FormatUserUID(api.SystemBotID)
	for _, binding := range policy.GetBindings() {
		if binding.GetRole() != workspaceAdminRole {
			continue
		}
		for _, member := range binding.GetMembers() {
			if member != userMember && member != systemBotMember {
				return true
			}
		}
	}
	return false
}

// UndeleteUser undeletes a user.
func (s *AuthService) UndeleteUser(ctx context.Context, request *v1pb.UndeleteUserRequest) (*v1pb.User, error) {
	if err := s.userCountGuard(ctx); err != nil {
		return nil, err
	}

	callerUser, ok := ctx.Value(common.UserContextKey).(*store.UserMessage)
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "failed to get caller user")
	}
	ok, err := s.iamManager.CheckPermission(ctx, iam.PermissionUsersUndelete, callerUser)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "user does not have permission %q", iam.PermissionUsersUndelete)
	}

	userID, err := common.GetUserID(request.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user, error: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user %d not found", userID)
	}
	if !user.MemberDeleted {
		return nil, status.Errorf(codes.InvalidArgument, "user %q is already active", userID)
	}

	user, err = s.store.UpdateUser(ctx, user, &store.UpdateUserMessage{Delete: &undeletePatch}, callerUser.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return convertToUser(user), nil
}

func convertToUser(user *store.UserMessage) *v1pb.User {
	userType := v1pb.UserType_USER_TYPE_UNSPECIFIED
	switch user.Type {
	case api.EndUser:
		userType = v1pb.UserType_USER
	case api.SystemBot:
		userType = v1pb.UserType_SYSTEM_BOT
	case api.ServiceAccount:
		userType = v1pb.UserType_SERVICE_ACCOUNT
	}

	convertedUser := &v1pb.User{
		Name:     common.FormatUserUID(user.ID),
		State:    convertDeletedToState(user.MemberDeleted),
		Email:    user.Email,
		Phone:    user.Phone,
		Title:    user.Name,
		UserType: userType,
		Profile: &v1pb.User_Profile{
			LastLoginTime:          user.Profile.LastLoginTime,
			LastChangePasswordTime: user.Profile.LastChangePasswordTime,
			Source:                 user.Profile.Source,
		},
	}

	if user.MFAConfig != nil {
		convertedUser.MfaEnabled = user.MFAConfig.OtpSecret != ""
		convertedUser.MfaSecret = user.MFAConfig.TempOtpSecret
		convertedUser.RecoveryCodes = user.MFAConfig.TempRecoveryCodes
	}
	return convertedUser
}

func convertToPrincipalType(userType v1pb.UserType) (api.PrincipalType, error) {
	var t api.PrincipalType
	switch userType {
	case v1pb.UserType_USER:
		t = api.EndUser
	case v1pb.UserType_SYSTEM_BOT:
		t = api.SystemBot
	case v1pb.UserType_SERVICE_ACCOUNT:
		t = api.ServiceAccount
	default:
		return t, status.Errorf(codes.InvalidArgument, "invalid user type %s", userType)
	}
	return t, nil
}

func (s *AuthService) needResetPassword(ctx context.Context, user *store.UserMessage) bool {
	// Reset password restriction only works for end user with email & password login.
	if user.Type != api.EndUser {
		return false
	}
	if err := s.licenseService.IsFeatureEnabled(api.FeaturePasswordRestriction); err != nil {
		return false
	}

	passwordRestriction, err := s.store.GetPasswordRestrictionSetting(ctx)
	if err != nil {
		slog.Error("failed to get password restriction", log.BBError(err))
		return false
	}

	if user.Profile.LastLoginTime == nil {
		if !passwordRestriction.RequireResetPasswordForFirstLogin {
			return false
		}
		count, err := s.store.CountUsers(ctx, api.EndUser)
		if err != nil {
			slog.Error("failed to count end users", log.BBError(err))
			return false
		}
		// The 1st workspace admin login don't need to reset the password
		return count > 1
	}

	if passwordRestriction.PasswordRotation != nil && passwordRestriction.PasswordRotation.GetNanos() > 0 {
		lastChangePasswordTime := user.CreatedTime
		if user.Profile.LastChangePasswordTime != nil {
			lastChangePasswordTime = user.Profile.LastChangePasswordTime.AsTime()
		}
		if lastChangePasswordTime.Add(time.Duration(passwordRestriction.PasswordRotation.GetNanos())).Before(time.Now()) {
			return true
		}
	}

	return false
}

// Login is the auth login method including SSO.
func (s *AuthService) Login(ctx context.Context, request *v1pb.LoginRequest) (*v1pb.LoginResponse, error) {
	var loginUser *store.UserMessage
	mfaSecondLogin := false
	if request.MfaTempToken != nil && *request.MfaTempToken != "" {
		mfaSecondLogin = true
	}
	loginViaIDP := request.GetIdpName() != ""

	response := &v1pb.LoginResponse{}
	if !mfaSecondLogin {
		var err error
		if loginViaIDP {
			loginUser, err = s.getOrCreateUserWithIDP(ctx, request)
			if err != nil {
				return nil, err
			}
		} else {
			loginUser, err = s.getAndVerifyUser(ctx, request)
			if err != nil {
				return nil, err
			}
			// Reset password restriction only works for end user with email & password login.
			response.RequireResetPassword = s.needResetPassword(ctx, loginUser)
		}
	} else {
		userID, err := auth.GetUserIDFromMFATempToken(*request.MfaTempToken, s.profile.Mode, s.secret)
		if err != nil {
			return nil, err
		}
		user, err := s.store.GetUserByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, invalidUserOrPasswordError
		}

		if request.OtpCode != nil {
			if err := challengeMFACode(user, *request.OtpCode); err != nil {
				return nil, err
			}
		} else if request.RecoveryCode != nil {
			if err := s.challengeRecoveryCode(ctx, user, *request.RecoveryCode); err != nil {
				return nil, err
			}
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "OTP or recovery code is required for MFA")
		}
		loginUser = user
	}

	if loginUser.MemberDeleted {
		return nil, status.Errorf(codes.Unauthenticated, "user has been deactivated by administrators")
	}

	setting, err := s.store.GetWorkspaceGeneralSetting(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find workspace setting, error: %v", err)
	}
	isWorkspaceAdmin, err := s.isUserWorkspaceAdmin(ctx, loginUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check user roles, error: %v", err)
	}
	if !isWorkspaceAdmin && loginUser.Type == api.EndUser && !mfaSecondLogin {
		// Disallow password signin for end users.
		if setting.DisallowPasswordSignin && !loginViaIDP {
			return nil, status.Errorf(codes.PermissionDenied, "password signin is disallowed")
		}
		// Check domain restriction for end users.
		var allowedDomains []string
		if setting.EnforceIdentityDomain {
			allowedDomains = setting.Domains
		}
		if err := validateEmail(loginUser.Email, allowedDomains, false); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid email %q, error: %v", loginUser.Email, err)
		}
	}

	tokenDuration := auth.GetTokenDuration(ctx, s.store)
	userMFAEnabled := loginUser.MFAConfig != nil && loginUser.MFAConfig.OtpSecret != ""
	// We only allow MFA login (2-step) when the feature is enabled and user has enabled MFA.
	if s.licenseService.IsFeatureEnabled(api.Feature2FA) == nil && !mfaSecondLogin && userMFAEnabled {
		mfaTempToken, err := auth.GenerateMFATempToken(loginUser.Name, loginUser.ID, s.profile.Mode, s.secret, tokenDuration)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate MFA temp token")
		}
		return &v1pb.LoginResponse{
			MfaTempToken: &mfaTempToken,
		}, nil
	}

	if loginUser.Type == api.EndUser {
		token, err := auth.GenerateAccessToken(loginUser.Name, loginUser.ID, s.profile.Mode, s.secret, tokenDuration)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate API access token")
		}
		response.Token = token
	} else if loginUser.Type == api.ServiceAccount {
		token, err := auth.GenerateAPIToken(loginUser.Name, loginUser.ID, s.profile.Mode, s.secret)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate API access token")
		}
		response.Token = token
	} else {
		return nil, status.Errorf(codes.Unauthenticated, "user type %s cannot login", loginUser.Type)
	}

	if request.Web {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "failed to parse metadata from incoming context")
		}
		// Pass the request origin header to response.
		var origin string
		for _, v := range md.Get("grpcgateway-origin") {
			origin = v
		}
		if err := grpc.SetHeader(ctx, metadata.New(map[string]string{
			auth.GatewayMetadataAccessTokenKey:   response.Token,
			auth.GatewayMetadataUserIDKey:        fmt.Sprintf("%d", loginUser.ID),
			auth.GatewayMetadataRequestOriginKey: origin,
		})); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to set grpc header, error: %v", err)
		}
	}

	if _, err := s.store.UpdateUser(ctx, loginUser, &store.UpdateUserMessage{
		Profile: &storepb.UserProfile{
			LastLoginTime:          timestamppb.Now(),
			LastChangePasswordTime: loginUser.Profile.GetLastChangePasswordTime(),
		},
	}, api.SystemBotID); err != nil {
		slog.Error("failed to update user profile", log.BBError(err), slog.String("user", loginUser.Email))
	}

	response.User = convertToUser(loginUser)

	s.metricReporter.Report(ctx, &metric.Metric{
		Name:  metricapi.PrincipalLoginMetricName,
		Value: 1,
		Labels: map[string]any{
			"email": loginUser.Email,
		},
	})
	return response, nil
}

// Logout is the auth logout method.
func (s *AuthService) Logout(ctx context.Context, _ *v1pb.LogoutRequest) (*emptypb.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "failed to parse metadata from incoming context")
	}
	accessTokenStr, err := auth.GetTokenFromMetadata(md)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	s.stateCfg.ExpireCache.Add(accessTokenStr, true)

	if err := grpc.SetHeader(ctx, metadata.New(map[string]string{
		auth.GatewayMetadataAccessTokenKey: "",
		auth.GatewayMetadataUserIDKey:      "",
	})); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set grpc header, error: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *AuthService) getAndVerifyUser(ctx context.Context, request *v1pb.LoginRequest) (*store.UserMessage, error) {
	user, err := s.store.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user by email %q: %v", request.Email, err)
	}
	if user == nil {
		return nil, invalidUserOrPasswordError
	}
	// Compare the stored hashed password, with the hashed version of the password that was received.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		// If the two passwords don't match, return a 401 status.
		return nil, invalidUserOrPasswordError
	}
	return user, nil
}

func (s *AuthService) getOrCreateUserWithIDP(ctx context.Context, request *v1pb.LoginRequest) (*store.UserMessage, error) {
	idpID, err := common.GetIdentityProviderID(request.IdpName)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to get identity provider ID: %v", err)
	}
	idp, err := s.store.GetIdentityProvider(ctx, &store.FindIdentityProviderMessage{
		ResourceID: &idpID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get identity provider: %v", err)
	}
	if idp == nil {
		return nil, status.Errorf(codes.NotFound, "identity provider not found")
	}

	setting, err := s.store.GetWorkspaceGeneralSetting(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get workspace setting: %v", err)
	}

	var userInfo *storepb.IdentityProviderUserInfo
	if idp.Type == storepb.IdentityProviderType_OAUTH2 {
		oauth2Context := request.IdpContext.GetOauth2Context()
		if oauth2Context == nil {
			return nil, status.Errorf(codes.InvalidArgument, "missing OAuth2 context")
		}
		oauth2IdentityProvider, err := oauth2.NewIdentityProvider(idp.Config.GetOauth2Config())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create new OAuth2 identity provider: %v", err)
		}
		redirectURL := fmt.Sprintf("%s/oauth/callback", setting.ExternalUrl)
		token, err := oauth2IdentityProvider.ExchangeToken(ctx, redirectURL, oauth2Context.Code)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to exchange token: %v", err)
		}
		userInfo, err = oauth2IdentityProvider.UserInfo(token)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get user info: %v", err)
		}
	} else if idp.Type == storepb.IdentityProviderType_OIDC {
		oauth2Context := request.IdpContext.GetOauth2Context()
		if oauth2Context == nil {
			return nil, status.Errorf(codes.InvalidArgument, "missing OAuth2 context")
		}

		idpConfig := idp.Config.GetOidcConfig()
		oidcIDP, err := oidc.NewIdentityProvider(
			ctx,
			oidc.IdentityProviderConfig{
				Issuer:        idpConfig.Issuer,
				ClientID:      idpConfig.ClientId,
				ClientSecret:  idpConfig.ClientSecret,
				FieldMapping:  idpConfig.FieldMapping,
				SkipTLSVerify: idpConfig.SkipTlsVerify,
				AuthStyle:     idpConfig.GetAuthStyle(),
			},
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create new OIDC identity provider: %v", err)
		}

		redirectURL := fmt.Sprintf("%s/oidc/callback", setting.ExternalUrl)
		token, err := oidcIDP.ExchangeToken(ctx, redirectURL, oauth2Context.Code)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to exchange token: %v", err)
		}

		userInfo, err = oidcIDP.UserInfo(ctx, token, "")
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get user info: %v", err)
		}
	} else if idp.Type == storepb.IdentityProviderType_LDAP {
		idpConfig := idp.Config.GetLdapConfig()
		ldapIDP, err := ldap.NewIdentityProvider(
			ldap.IdentityProviderConfig{
				Host:             idpConfig.Host,
				Port:             int(idpConfig.Port),
				SkipTLSVerify:    idpConfig.SkipTlsVerify,
				BindDN:           idpConfig.BindDn,
				BindPassword:     idpConfig.BindPassword,
				BaseDN:           idpConfig.BaseDn,
				UserFilter:       idpConfig.UserFilter,
				SecurityProtocol: ldap.SecurityProtocol(idpConfig.SecurityProtocol),
				FieldMapping:     idpConfig.FieldMapping,
			},
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create new LDAP identity provider: %v", err)
		}

		userInfo, err = ldapIDP.Authenticate(request.Email, request.Password)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get user info: %v", err)
		}
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "identity provider type %s not supported", idp.Type.String())
	}
	if userInfo == nil {
		return nil, status.Errorf(codes.NotFound, "identity provider user info not found")
	}

	// The userinfo's email comes from identity provider, it has to be converted to lower-case.
	email := strings.ToLower(userInfo.Identifier)
	var allowedDomains []string
	if setting.EnforceIdentityDomain {
		allowedDomains = setting.Domains
	}
	if err := validateEmail(email, allowedDomains, false /* isServiceAccount */); err != nil {
		// If the email is invalid, we will try to use the domain and identifier to construct the email.
		if idp.Domain != "" {
			domain := extractDomain(idp.Domain)
			email = strings.ToLower(fmt.Sprintf("%s@%s", userInfo.Identifier, domain))
		}
	}

	// If the email is still invalid, we will return an error.
	if err := validateEmail(email, allowedDomains, false /* isServiceAccount */); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email %q, error: %v", email, err)
	}
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users by email %s: %v", email, err)
	}
	if user != nil {
		if user.MemberDeleted {
			// Undelete the user when login via SSO.
			user, err = s.store.UpdateUser(ctx, user, &store.UpdateUserMessage{Delete: &undeletePatch}, api.SystemBotID)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to undelete user: %v", err)
			}
		}
		return user, nil
	}

	if err := s.licenseService.IsFeatureEnabled(api.FeatureSSO); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	// Create new user from identity provider.
	password, err := common.RandomString(20)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate random password")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate password hash")
	}
	newUser, err := s.store.CreateUser(ctx, &store.UserMessage{
		Name:         userInfo.DisplayName,
		Email:        email,
		Phone:        userInfo.Phone,
		Type:         api.EndUser,
		PasswordHash: string(passwordHash),
	}, api.SystemBotID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user, error: %v", err)
	}
	return newUser, nil
}

func challengeMFACode(user *store.UserMessage, mfaCode string) error {
	if !validateWithCodeAndSecret(mfaCode, user.MFAConfig.OtpSecret) {
		return status.Errorf(codes.Unauthenticated, "invalid MFA code")
	}
	return nil
}

func (s *AuthService) challengeRecoveryCode(ctx context.Context, user *store.UserMessage, recoveryCode string) error {
	for i, code := range user.MFAConfig.RecoveryCodes {
		if code == recoveryCode {
			// If the recovery code is valid, delete it from the user's recovery code list.
			user.MFAConfig.RecoveryCodes = slices.Delete(user.MFAConfig.RecoveryCodes, i, i+1)
			_, err := s.store.UpdateUser(ctx, user, &store.UpdateUserMessage{
				MFAConfig: &storepb.MFAConfig{
					OtpSecret:     user.MFAConfig.OtpSecret,
					RecoveryCodes: user.MFAConfig.RecoveryCodes,
				},
			}, user.ID)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to update user: %v", err)
			}
			return nil
		}
	}
	return status.Errorf(codes.Unauthenticated, "invalid recovery code")
}

func validateEmail(email string, allowedDomains []string, isServiceAccount bool) error {
	formattedEmail := strings.ToLower(email)
	if email != formattedEmail {
		return errors.New("email should be lowercase")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return err
	}
	// Domain restrictions are not applied to service account.
	if isServiceAccount {
		return nil
	}
	// Enforce domain restrictions.
	if len(allowedDomains) > 0 {
		ok := false
		for _, v := range allowedDomains {
			if strings.HasSuffix(email, fmt.Sprintf("@%s", v)) {
				ok = true
				break
			}
		}
		if !ok {
			return errors.Errorf("email %q does not belong to domains %v", email, allowedDomains)
		}
	}
	return nil
}

func extractDomain(input string) string {
	pattern := `[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+`
	regExp, err := regexp.Compile(pattern)
	if err != nil {
		// WHen the pattern is invalid, we just return the input.
		return input
	}

	match := regExp.FindString(input)
	domainParts := strings.Split(match, ".")
	// If the domain has at least 3 parts, we will remove the first part.
	if len(domainParts) >= 3 {
		match = strings.Join(domainParts[1:], ".")
	}
	return match
}

const (
	// issuerName is the name of the issuer of the OTP token.
	issuerName = "Devsecdb"
)

// generateRandSecret generates a random secret for the given account name.
func generateRandSecret(accountName string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuerName,
		AccountName: accountName,
	})
	if err != nil {
		return "", err
	}
	return key.Secret(), nil
}

// validateWithCodeAndSecret validates the given code against the given secret.
func validateWithCodeAndSecret(code, secret string) bool {
	return totp.Validate(code, secret)
}

// generateRecoveryCodes generates n recovery codes.
func generateRecoveryCodes(n int) ([]string, error) {
	recoveryCodes := make([]string, n)
	for i := 0; i < n; i++ {
		code, err := common.RandomString(10)
		if err != nil {
			return nil, err
		}
		recoveryCodes[i] = code
	}
	return recoveryCodes, nil
}

func (s *AuthService) userCountGuard(ctx context.Context) error {
	userLimit := s.licenseService.GetPlanLimitValue(ctx, enterprise.PlanLimitMaximumUser)

	count, err := s.store.CountActiveUsers(ctx)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	if count >= userLimit {
		return status.Errorf(codes.ResourceExhausted, "reached the maximum user count %d", userLimit)
	}
	return nil
}

func (s *AuthService) isUserWorkspaceAdmin(ctx context.Context, user *store.UserMessage) (bool, error) {
	workspacePolicy, err := s.store.GetWorkspaceIamPolicy(ctx)
	if err != nil {
		return false, err
	}
	roles := utils.GetUserFormattedRolesMap(ctx, s.store, user, workspacePolicy.Policy)
	return roles[common.FormatRole(api.WorkspaceAdmin.String())], nil
}
