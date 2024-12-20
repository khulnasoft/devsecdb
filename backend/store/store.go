// Package store is the implementation for managing Devsecdb's own metadata in a PostgreSQL database.
package store

import (
	"context"
	"fmt"
	"strings"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/khulnasoft/devsecdb/backend/component/config"
	api "github.com/khulnasoft/devsecdb/backend/legacyapi"
	"github.com/khulnasoft/devsecdb/backend/store/model"
)

// Store provides database access to all raw objects.
type Store struct {
	db      *DB
	profile *config.Profile

	userIDCache            *lru.Cache[int, *UserMessage]
	userEmailCache         *lru.Cache[string, *UserMessage]
	environmentCache       *lru.Cache[string, *EnvironmentMessage]
	environmentIDCache     *lru.Cache[int, *EnvironmentMessage]
	instanceCache          *lru.Cache[string, *InstanceMessage]
	instanceIDCache        *lru.Cache[int, *InstanceMessage]
	databaseCache          *lru.Cache[string, *DatabaseMessage]
	databaseIDCache        *lru.Cache[int, *DatabaseMessage]
	projectCache           *lru.Cache[string, *ProjectMessage]
	projectIDCache         *lru.Cache[int, *ProjectMessage]
	projectDeploymentCache *lru.Cache[int, *DeploymentConfigMessage]
	policyCache            *lru.Cache[string, *PolicyMessage]
	issueCache             *lru.Cache[int, *IssueMessage]
	issueByPipelineCache   *lru.Cache[int, *IssueMessage]
	pipelineCache          *lru.Cache[int, *PipelineMessage]
	settingCache           *lru.Cache[api.SettingName, *SettingMessage]
	idpCache               *lru.Cache[string, *IdentityProviderMessage]
	risksCache             *lru.Cache[int, []*RiskMessage] // Use 0 as the key.
	databaseGroupCache     *lru.Cache[string, *DatabaseGroupMessage]
	databaseGroupIDCache   *lru.Cache[int64, *DatabaseGroupMessage]
	vcsIDCache             *lru.Cache[int, *VCSProviderMessage]
	rolesCache             *lru.Cache[string, *RoleMessage]
	groupCache             *lru.Cache[string, *GroupMessage]
	sheetCache             *lru.Cache[int, *SheetMessage]

	// Large objects.
	sheetStatementCache *lru.Cache[int, string]
	dbSchemaCache       *lru.Cache[int, *model.DBSchema]
}

// New creates a new instance of Store.
func New(db *DB, profile *config.Profile) (*Store, error) {
	userIDCache, err := lru.New[int, *UserMessage](32768)
	if err != nil {
		return nil, err
	}
	userEmailCache, err := lru.New[string, *UserMessage](32768)
	if err != nil {
		return nil, err
	}
	environmentCache, err := lru.New[string, *EnvironmentMessage](32)
	if err != nil {
		return nil, err
	}
	environmentIDCache, err := lru.New[int, *EnvironmentMessage](32)
	if err != nil {
		return nil, err
	}
	instanceCache, err := lru.New[string, *InstanceMessage](32768)
	if err != nil {
		return nil, err
	}
	instanceIDCache, err := lru.New[int, *InstanceMessage](32768)
	if err != nil {
		return nil, err
	}
	databaseCache, err := lru.New[string, *DatabaseMessage](32768)
	if err != nil {
		return nil, err
	}
	databaseIDCache, err := lru.New[int, *DatabaseMessage](32768)
	if err != nil {
		return nil, err
	}
	projectCache, err := lru.New[string, *ProjectMessage](32768)
	if err != nil {
		return nil, err
	}
	projectIDCache, err := lru.New[int, *ProjectMessage](32768)
	if err != nil {
		return nil, err
	}
	projectDeploymentCache, err := lru.New[int, *DeploymentConfigMessage](128)
	if err != nil {
		return nil, err
	}
	policyCache, err := lru.New[string, *PolicyMessage](128)
	if err != nil {
		return nil, err
	}
	issueCache, err := lru.New[int, *IssueMessage](256)
	if err != nil {
		return nil, err
	}
	issueByPipelineCache, err := lru.New[int, *IssueMessage](256)
	if err != nil {
		return nil, err
	}
	pipelineCache, err := lru.New[int, *PipelineMessage](256)
	if err != nil {
		return nil, err
	}
	settingCache, err := lru.New[api.SettingName, *SettingMessage](64)
	if err != nil {
		return nil, err
	}
	idpCache, err := lru.New[string, *IdentityProviderMessage](4)
	if err != nil {
		return nil, err
	}
	risksCache, err := lru.New[int, []*RiskMessage](4)
	if err != nil {
		return nil, err
	}
	databaseGroupCache, err := lru.New[string, *DatabaseGroupMessage](1024)
	if err != nil {
		return nil, err
	}
	databaseGroupIDCache, err := lru.New[int64, *DatabaseGroupMessage](1024)
	if err != nil {
		return nil, err
	}
	vcsIDCache, err := lru.New[int, *VCSProviderMessage](1024)
	if err != nil {
		return nil, err
	}
	rolesCache, err := lru.New[string, *RoleMessage](64)
	if err != nil {
		return nil, err
	}
	sheetCache, err := lru.New[int, *SheetMessage](64)
	if err != nil {
		return nil, err
	}
	sheetStatementCache, err := lru.New[int, string](10)
	if err != nil {
		return nil, err
	}
	dbSchemaCache, err := lru.New[int, *model.DBSchema](128)
	if err != nil {
		return nil, err
	}
	groupCache, err := lru.New[string, *GroupMessage](1024)
	if err != nil {
		return nil, err
	}

	return &Store{
		db:      db,
		profile: profile,

		// Cache.
		userIDCache:            userIDCache,
		userEmailCache:         userEmailCache,
		environmentCache:       environmentCache,
		environmentIDCache:     environmentIDCache,
		instanceCache:          instanceCache,
		instanceIDCache:        instanceIDCache,
		databaseCache:          databaseCache,
		databaseIDCache:        databaseIDCache,
		projectCache:           projectCache,
		projectIDCache:         projectIDCache,
		projectDeploymentCache: projectDeploymentCache,
		policyCache:            policyCache,
		issueCache:             issueCache,
		issueByPipelineCache:   issueByPipelineCache,
		pipelineCache:          pipelineCache,
		settingCache:           settingCache,
		idpCache:               idpCache,
		risksCache:             risksCache,
		databaseGroupCache:     databaseGroupCache,
		databaseGroupIDCache:   databaseGroupIDCache,
		vcsIDCache:             vcsIDCache,
		rolesCache:             rolesCache,
		sheetCache:             sheetCache,
		sheetStatementCache:    sheetStatementCache,
		dbSchemaCache:          dbSchemaCache,
		groupCache:             groupCache,
	}, nil
}

// Close closes underlying db.
func (s *Store) Close(ctx context.Context) error {
	return s.db.Close(ctx)
}

func getInstanceCacheKey(instanceID string) string {
	return instanceID
}

func getPolicyCacheKey(resourceType api.PolicyResourceType, resourceUID int, policyType api.PolicyType) string {
	return fmt.Sprintf("policies/%s/%d/%s", resourceType, resourceUID, policyType)
}

func getDatabaseCacheKey(instanceID, databaseName string) string {
	return fmt.Sprintf("%s/%s", instanceID, databaseName)
}

func getDatabaseGroupCacheKey(projectUID int, databaseGroupResourceID string) string {
	return fmt.Sprintf("%d/%s", projectUID, databaseGroupResourceID)
}

func getPlaceholders(start int, count int) string {
	var placeholders []string
	for i := start; i < start+count; i++ {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
	}
	return fmt.Sprintf("(%s)", strings.Join(placeholders, ","))
}
