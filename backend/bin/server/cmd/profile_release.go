//go:build release

package cmd

import (
	"time"

	"github.com/khulnasoft/devsecdb/backend/common"
	"github.com/khulnasoft/devsecdb/backend/component/config"
)

func activeProfile(dataDir string) *config.Profile {
	p := getBaseProfile(dataDir)
	p.RuntimeDebug.Store(p.Debug)
	p.Mode = common.ReleaseModeProd
	p.PgUser = "bb"
	p.BackupRunnerInterval = 10 * time.Minute
	p.AppRunnerInterval = 30 * time.Second
	// Enable metric if it's not explicitly disabled and it's not running in demo mode.
	p.EnableMetric = !flags.disableMetric && p.DemoName == ""
	p.MetricConnectionKey = "so9lLwj5zLjH09sxNabsyVNYSsAHn68F"
	return p
}
