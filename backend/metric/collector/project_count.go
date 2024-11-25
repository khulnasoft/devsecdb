package collector

import (
	"context"

	metricapi "github.com/khulnasoft/devsecdb/backend/metric"
	"github.com/khulnasoft/devsecdb/backend/plugin/metric"
	"github.com/khulnasoft/devsecdb/backend/store"
)

var _ metric.Collector = (*projectCountCollector)(nil)

// projectCountCollector is the metric data collector for project.
type projectCountCollector struct {
	store *store.Store
}

// NewProjectCountCollector creates a new instance of projectCollector.
func NewProjectCountCollector(store *store.Store) metric.Collector {
	return &projectCountCollector{
		store: store,
	}
}

// Collect will collect the metric for project.
func (c *projectCountCollector) Collect(ctx context.Context) ([]*metric.Metric, error) {
	var res []*metric.Metric

	projectCountMetricList, err := c.store.CountProjectGroupByWorkflow(ctx)
	if err != nil {
		return nil, err
	}

	for _, projectCountMetric := range projectCountMetricList {
		res = append(res, &metric.Metric{
			Name:  metricapi.ProjectCountMetricName,
			Value: projectCountMetric.Count,
			Labels: map[string]any{
				"workflow": projectCountMetric.WorkflowType.String(),
				"status":   string(projectCountMetric.RowStatus),
			},
		})
	}

	return res, nil
}
