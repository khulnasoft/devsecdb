package directorysync

import (
	"github.com/khulnasoft/devsecdb/backend/component/iam"
	enterprise "github.com/khulnasoft/devsecdb/backend/enterprise/api"
	"github.com/khulnasoft/devsecdb/backend/store"
)

// Service is the API endpoint for handling SCIM requests.
type Service struct {
	store          *store.Store
	licenseService enterprise.LicenseService
	iamManager     *iam.Manager
}

// NewService creates a SCIM service.
func NewService(
	store *store.Store,
	licenseService enterprise.LicenseService,
	iamManager *iam.Manager,
) *Service {
	return &Service{
		store:          store,
		licenseService: licenseService,
		iamManager:     iamManager,
	}
}
