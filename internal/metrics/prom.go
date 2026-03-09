package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusCollector struct {
	backupsTotal          *prometheus.CounterVec
	backupsCreateDuration *prometheus.HistogramVec
}

func NewPrometheusCollector(namespace string) *PrometheusCollector {
	return &PrometheusCollector{
		backupsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "backups_total",
		}, []string{"database"}),
		backupsCreateDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "backups_create_duration_seconds",
		}, []string{"database"}),
	}
}

func (c *PrometheusCollector) IncBackupsTotal(dbName string) {
	c.backupsTotal.WithLabelValues(dbName).Inc()
}

func (c *PrometheusCollector) ObserveBackupCreateDuration(dbName string, dur time.Duration) {
	c.backupsCreateDuration.WithLabelValues(dbName).Observe(dur.Seconds())
}

func (c *PrometheusCollector) Flush(_ context.Context) error {
	return nil
}

func (c *PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	c.backupsTotal.Describe(ch)
	c.backupsCreateDuration.Describe(ch)
}

func (c *PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	c.backupsTotal.Collect(ch)
	c.backupsCreateDuration.Collect(ch)
}
