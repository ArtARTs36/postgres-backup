package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/artarts36/postgres-backup/internal/backup"
	"github.com/artarts36/postgres-backup/internal/config"
	"github.com/artarts36/postgres-backup/internal/metrics"
	"github.com/artarts36/postgres-backup/internal/storage"
)

type App struct {
	cfg *config.Config

	backupCreator *backup.Creator
	backupCleaner *backup.Cleaner

	storage storage.Storage
}

func NewApp(cfg *config.Config, metrics metrics.Collector) (*App, error) {
	app := &App{
		cfg: cfg,
	}

	if err := app.initStorage(); err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}

	var err error
	app.backupCreator, err = backup.NewCreator(metrics, app.storage)
	if err != nil {
		return nil, fmt.Errorf("init backup creator: %w", err)
	}

	app.backupCleaner = backup.NewCleaner(app.storage)

	return app, nil
}

func (app *App) Run(ctx context.Context) error {
	var errs []error

	for _, dbName := range app.cfg.Postgres.Database {
		slog.InfoContext(ctx, "[app] start backup", slog.String("database", dbName))

		err := app.run(ctx, dbName)
		if err != nil {
			slog.ErrorContext(ctx, "[app] failed to backup", slog.String("database", dbName))

			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (app *App) run(ctx context.Context, dbName string) error {
	err := app.backupCreator.Create(ctx, backup.CreatingBackup{
		Host:     app.cfg.Postgres.Host,
		Port:     app.cfg.Postgres.Port,
		User:     app.cfg.Postgres.User,
		Password: app.cfg.Postgres.Password,
		Database: dbName,
		TempDir:  app.cfg.TempDir,
	})
	if err != nil {
		return fmt.Errorf("create backup: %w", err)
	}

	slog.Info("[app] backup created successfully",
		slog.String("database", dbName),
	)

	slog.Info("[app] cleaning old backups",
		slog.String("database", dbName),
	)

	err = app.backupCleaner.CleanOldBackups(ctx, dbName, app.cfg.MaxBackups)
	if err != nil {
		return fmt.Errorf("cleaning old backups: %w", err)
	}

	return nil
}

func (app *App) Close() error {
	if app.storage != nil {
		return app.storage.Close()
	}

	return nil
}

func (app *App) initStorage() error {
	var (
		store storage.Storage
		err   error
	)

	switch app.cfg.StorageType {
	case config.StorageTypeFS:
		if app.cfg.FSRoot == "" {
			slog.Error("FS_ROOT is required when STORAGE_TYPE=fs")
			os.Exit(1)
		}
		store, err = storage.NewFileSystemStorage(app.cfg.FSRoot)
		if err != nil {
			return fmt.Errorf("init filesystem storage: %w", err)
		}
	case config.StorageTypeS3:
		store, err = storage.NewS3Storage(app.cfg.S3)
		if err != nil {
			return fmt.Errorf("init s3 storage: %w", err)
		}
	default:
		return fmt.Errorf("unknown storage type %q", app.cfg.StorageType)
	}

	app.storage = store

	return nil
}
