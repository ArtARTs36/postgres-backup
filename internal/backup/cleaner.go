package backup

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"sort"
	"time"

	"github.com/artarts36/postgres-backup/internal/storage"
)

type Cleaner struct {
	storage storage.Storage
}

func NewCleaner(storage storage.Storage) *Cleaner {
	return &Cleaner{storage: storage}
}

func (bc *Cleaner) CleanOldBackups(ctx context.Context, dbName string, maxBackups int) error {
	if maxBackups <= 0 {
		return nil
	}

	prefix := dbName + "/"
	objects, err := bc.storage.List(prefix)
	if err != nil {
		return fmt.Errorf("list objects: %w", err)
	}

	var backups []struct {
		Key  string
		Time time.Time
	}

	for _, obj := range objects {
		filename := path.Base(obj.Key)
		matches := backupFileRegex.FindStringSubmatch(filename)
		if len(matches) != 2 { //nolint:mnd // not need
			continue
		}

		timestampStr := matches[1]
		fileTime, terr := time.Parse("2006-01-02T15-04-05Z", timestampStr)
		if terr != nil {
			continue
		}

		backups = append(backups, struct {
			Key  string
			Time time.Time
		}{Key: obj.Key, Time: fileTime})
	}

	if len(backups) <= maxBackups {
		return nil
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Time.Before(backups[j].Time)
	})

	deleted := 0
	fails := 0

	for i := 0; i < len(backups)-maxBackups; i++ {
		if err = bc.storage.Delete(ctx, backups[i].Key); err != nil {
			slog.WarnContext(ctx, "[s3-storage] failed to delete object",
				slog.String("key", backups[i].Key),
				slog.Any("err", err),
			)
			fails++
		} else {
			deleted++
		}
	}

	slog.InfoContext(ctx, "cleaned old backups",
		slog.Int("deleted", deleted),
		slog.Int("fails", fails),
	)
	return nil
}
