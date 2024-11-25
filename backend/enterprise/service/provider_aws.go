//go:build aws

package service

import (
	"github.com/khulnasoft/devsecdb/backend/enterprise/plugin"
	"github.com/khulnasoft/devsecdb/backend/enterprise/plugin/aws"
)

func getLicenseProvider(providerConfig *plugin.ProviderConfig) (plugin.LicenseProvider, error) {
	return aws.NewProvider(providerConfig)
}
