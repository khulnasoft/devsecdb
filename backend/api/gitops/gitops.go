// Package gitops is the package for GitOps APIs.
package gitops

import (
	v1pb "github.com/khulnasoft/devsecdb/backend/api/v1"
	"github.com/khulnasoft/devsecdb/backend/component/config"
	"github.com/khulnasoft/devsecdb/backend/component/sheet"
	enterprise "github.com/khulnasoft/devsecdb/backend/enterprise/api"
	"github.com/khulnasoft/devsecdb/backend/store"
)

// Service is the API endpoint for handling GitOps requests.
type Service struct {
	store          *store.Store
	licenseService enterprise.LicenseService
	releaseService *v1pb.ReleaseService
	planService    *v1pb.PlanService
	rolloutService *v1pb.RolloutService
	issueService   *v1pb.IssueService
	sqlService     *v1pb.SQLService
	sheetManager   *sheet.Manager
	profile        *config.Profile
}

// NewService creates a GitOps service.
func NewService(
	store *store.Store,
	licenseService enterprise.LicenseService,
	releaseService *v1pb.ReleaseService,
	planService *v1pb.PlanService,
	rolloutService *v1pb.RolloutService,
	issueService *v1pb.IssueService,
	sqlService *v1pb.SQLService,
	sheetManager *sheet.Manager,
	profile *config.Profile,
) *Service {
	return &Service{
		store:          store,
		licenseService: licenseService,
		releaseService: releaseService,
		planService:    planService,
		rolloutService: rolloutService,
		issueService:   issueService,
		sqlService:     sqlService,
		sheetManager:   sheetManager,
		profile:        profile,
	}
}
