package plugin

import (
	"context"

	"github.com/khulnasoft/devsecdb/backend/common"
	enterprise "github.com/khulnasoft/devsecdb/backend/enterprise/api"
	"github.com/khulnasoft/devsecdb/backend/store"
)

// LicenseProvider is the license provider.
type LicenseProvider interface {
	// StoreLicense will store the license.
	StoreLicense(ctx context.Context, patch *enterprise.SubscriptionPatch) error
	// LoadSubscription will load the subscription.
	LoadSubscription(ctx context.Context) *enterprise.Subscription
}

// ProviderConfig is the provider configuration.
type ProviderConfig struct {
	Mode  common.ReleaseMode
	Store *store.Store
}
