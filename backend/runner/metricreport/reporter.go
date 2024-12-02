// Package metricreport is a runner reporting metrics.
package metricreport

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/khulnasoft/devsecdb/backend/common/log"
	"github.com/khulnasoft/devsecdb/backend/component/config"
	enterprise "github.com/khulnasoft/devsecdb/backend/enterprise/api"
	api "github.com/khulnasoft/devsecdb/backend/legacyapi"
	"github.com/khulnasoft/devsecdb/backend/plugin/metric"
	"github.com/khulnasoft/devsecdb/backend/plugin/metric/segment"
	"github.com/khulnasoft/devsecdb/backend/store"
)

const (
	metricSchedulerInterval = 6 * time.Hour
	// identifyTraitForPlan is the trait key for subscription plan.
	identifyTraitForPlan = "plan"
	// identifyTraitForTrial is the trait key for trialing.
	identifyTraitForTrial = "trial"
	// identifyTraitForSubscriptionStartDate is the trait key for subscription start date.
	identifyTraitForSubscriptionStartDate = "subscription_start"
	// identifyTraitForSubscriptionEndDate is the trait key for subscription end date.
	identifyTraitForSubscriptionEndDate = "subscription_end"
	// identifyTraitForOrgID is the trait key for organization id.
	identifyTraitForOrgID = "org_id"
	// identifyTraitForOrgName is the trait key for organization name.
	identifyTraitForOrgName = "org_name"
	// identifyTraitForMode is the trait key for Devsecdb service mode.
	identifyTraitForMode = "mode"
	// identifyTraitForLastActiveTime is the trait key for Devsecdb last active time.
	identifyTraitForLastActiveTime = "last_active"
	// identifyTraitForVersion is the trait key for Devsecdb version.
	identifyTraitForVersion = "version"
	// devsecdbServiceModeSaaS is the mode for Devsecdb SaaS.
	bytebaseServiceModeSaaS = "saas"
	// devsecdbServiceModeSelfhost is the mode for Devsecdb self-host.
	bytebaseServiceModeSelfhost = "self-host"
)

// Reporter is the metric reporter.
type Reporter struct {
	licenseService enterprise.LicenseService
	profile        *config.Profile
	reporter       metric.Reporter
	collectors     map[string]metric.Collector
	store          *store.Store
}

// NewReporter creates a new metric scheduler.
func NewReporter(store *store.Store, licenseService enterprise.LicenseService, profile *config.Profile, enabled bool) *Reporter {
	var r metric.Reporter
	if enabled {
		r = segment.NewReporter(profile.MetricConnectionKey)
	} else {
		r = segment.NewMockReporter()
	}

	return &Reporter{
		licenseService: licenseService,
		profile:        profile,
		reporter:       r,
		collectors:     make(map[string]metric.Collector),
		store:          store,
	}
}

// Run will run the metric reporter.
func (m *Reporter) Run(ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(metricSchedulerInterval)
	defer ticker.Stop()
	defer wg.Done()

	slog.Debug(fmt.Sprintf("Metrics reporter started and will run every %v", metricSchedulerInterval))

	for {
		select {
		case <-ticker.C:
			go func() {
				defer func() {
					if r := recover(); r != nil {
						err, ok := r.(error)
						if !ok {
							err = errors.Errorf("%v", r)
						}
						slog.Error("Metrics reporter PANIC RECOVER", log.BBError(err), log.BBStack("panic-stack"))
					}
				}()

				ctx := context.Background()
				// identify will be triggered in every schedule loop so that we can update the latest workspace profile such as subscription plan.
				workspaceID, err := m.identify(ctx)
				if err != nil {
					slog.Error("failed to report identifier", log.BBError(err))
					return
				}

				for name, collector := range m.collectors {
					slog.Debug("Run metric collector", slog.String("collector", name))

					metricList, err := collector.Collect(ctx)
					if err != nil {
						slog.Error(
							"Failed to collect metric",
							slog.String("collector", name),
							log.BBError(err),
						)
						continue
					}

					for _, metric := range metricList {
						m.reportMetric(workspaceID, metric)
					}
				}
			}()
		case <-ctx.Done(): // if cancel() execute
			return
		}
	}
}

// Close will close the metric reporter.
func (m *Reporter) Close() {
	m.reporter.Close()
}

// Register will register a metric collector.
func (m *Reporter) Register(metricName metric.Name, collector metric.Collector) {
	m.collectors[string(metricName)] = collector
}

func (m *Reporter) reportMetric(id string, metric *metric.Metric) {
	if err := m.reporter.Report(id, metric); err != nil {
		slog.Error(
			"Failed to report metric",
			slog.String("metric", string(metric.Name)),
			log.BBError(err),
		)
	}
}

// Identify will identify the workspace and update the subscription plan.
func (m *Reporter) identify(ctx context.Context) (string, error) {
	workspaceID, err := m.store.GetWorkspaceID(ctx)
	if err != nil {
		return "", err
	}
	subscription := m.licenseService.LoadSubscription(ctx)
	plan := subscription.Plan.String()
	orgID := subscription.OrgID
	orgName := subscription.OrgName

	trial := "N"
	if subscription.Trialing {
		trial = "Y"
	}

	subscriptionStartDate := ""
	subscriptionEndDate := ""
	if subscription.Plan != api.FREE {
		subscriptionStartDate = time.Unix(subscription.StartedTs, 0).Format(time.RFC3339)
		subscriptionEndDate = time.Unix(subscription.ExpiresTs, 0).Format(time.RFC3339)
	}

	user, err := m.store.GetUserByID(ctx, api.PrincipalIDForFirstUser)
	if err != nil {
		slog.Debug("unable to get the first principal user", slog.Int("id", api.PrincipalIDForFirstUser), log.BBError(err))
	}
	email := ""
	name := ""
	if user != nil {
		email = user.Email
		name = user.Name
	}

	mode := devsecdbServiceModeSelfhost
	if m.profile.SaaS {
		mode = devsecdbServiceModeSaaS
	}

	if err := m.reporter.Identify(&metric.Identifier{
		ID:    workspaceID,
		Email: email,
		Name:  name,
		Labels: map[string]string{
			identifyTraitForPlan:                  plan,
			identifyTraitForTrial:                 trial,
			identifyTraitForVersion:               m.profile.Version,
			identifyTraitForOrgID:                 orgID,
			identifyTraitForOrgName:               orgName,
			identifyTraitForMode:                  mode,
			identifyTraitForLastActiveTime:        time.Unix(m.profile.LastActiveTs, 0).String(),
			identifyTraitForSubscriptionStartDate: subscriptionStartDate,
			identifyTraitForSubscriptionEndDate:   subscriptionEndDate,
		},
	}); err != nil {
		return workspaceID, err
	}

	return workspaceID, nil
}

// Report will report a metric.
func (m *Reporter) Report(ctx context.Context, metric *metric.Metric) {
	workspaceID, err := m.store.GetWorkspaceID(ctx)
	if err != nil {
		slog.Error("failed to find the workspace id", log.BBError(err))
		return
	}
	m.reportMetric(workspaceID, metric)
}
