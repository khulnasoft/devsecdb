//go:build !release

package cmd

import (
	"time"

	"github.com/khulnasoft/devsecdb/backend/common"
	"github.com/khulnasoft/devsecdb/backend/component/config"
)

func activeProfile(dataDir string) *config.Profile {
	p := getBaseProfile(dataDir)
	p.RuntimeDebug.Store(p.Debug)
	p.Mode = common.ReleaseModeDev
	p.PgUser = "bbdev"
	p.BackupRunnerInterval = 10 * time.Second
	p.AppRunnerInterval = 30 * time.Second
	p.EnableMetric = false
	// Metric collection is disabled in dev mode.
	// p.MetricConnectionKey = "3zcZLeX3ahvlueEJqNyJysGfVAErsjjT"
	return p
}
