package metrics

import (
	"context"
	"time"
)

type NoopCollector struct{}

func (n NoopCollector) IncBackupsTotal(_ string) {}

func (n NoopCollector) ObserveBackupCreateDuration(_ string, _ time.Duration) {}

func (n NoopCollector) Flush(_ context.Context) error { return nil }
