//go:build !aws

package service

import (
	"github.com/khulnasoft/devsecdb/backend/enterprise/plugin"
	"github.com/khulnasoft/devsecdb/backend/enterprise/plugin/hub"
)

func getLicenseProvider(providerConfig *plugin.ProviderConfig) (plugin.LicenseProvider, error) {
	return hub.NewProvider(providerConfig)
}
