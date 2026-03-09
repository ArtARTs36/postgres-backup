package metrics

import (
	"context"
	"time"
)

type Collector interface {
	IncBackupsTotal(dbName string)
	ObserveBackupCreateDuration(dbname string, dur time.Duration)
	Flush(ctx context.Context) error
}
