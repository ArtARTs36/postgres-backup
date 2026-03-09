package backup

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/artarts36/postgres-backup/internal/metrics"
	"github.com/artarts36/postgres-backup/internal/storage"
)

var (
	backupFileRegex = regexp.MustCompile(`^backup_(\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z)\.dump$`)
)

type CreatingBackup struct {
	Host     string
	Port     int
	User     string
	Password string //nolint:gosec // false-positive: no json
	Database string

	TempDir string
}

type Creator struct {
	metrics metrics.Collector
	storage storage.Storage
}

func getPgDumpVersion() (string, error) {
	const defaultTimeout = time.Second * 10

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pg_dump", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run 'pg_dump --version': %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func NewCreator(metrics metrics.Collector, store storage.Storage) (*Creator, error) {
	version, err := getPgDumpVersion()
	if err != nil {
		return nil, fmt.Errorf("pg_dump check failed: %w", err)
	}

	slog.Info("[creator] using pg_dump", slog.String("version", version))

	return &Creator{metrics: metrics, storage: store}, nil
}

func (bc *Creator) Create(ctx context.Context, req CreatingBackup) error {
	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05Z")
	tempFile := fmt.Sprintf("%s/backup_%s.dump", req.TempDir, timestamp)

	slog.Info("[creator] running pg_dump", slog.String("temp_file", tempFile))

	started := time.Now()

	cmd := exec.CommandContext(ctx, //nolint:gosec // not need
		"pg_dump",
		"-h", req.Host,
		"-p", fmt.Sprintf("%d", req.Port),
		"-U", req.User,
		"-d", req.Database,
		"-F", "c",
		"-f", tempFile,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", req.Password))
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run pg_dump: %w; stderr: %s", err, strings.TrimSpace(stderr.String()))
	}

	bc.metrics.IncBackupsTotal(req.Database)

	dur := time.Since(started)
	bc.metrics.ObserveBackupCreateDuration(req.Database, dur)

	slog.Info("[creator] dump saved", slog.String("file", tempFile))

	file, err := os.Open(tempFile)
	if err != nil {
		_ = os.Remove(tempFile)
		return fmt.Errorf("open dump file: %w", err)
	}

	key := fmt.Sprintf("%s/backup_%s.dump", req.Database, timestamp)
	if err = bc.storage.Put(key, file); err != nil {
		file.Close()
		_ = os.Remove(tempFile)
		return fmt.Errorf("upload failed: %w", err)
	}

	file.Close()
	_ = os.Remove(tempFile)

	slog.Info("[creator] upload completed")

	return nil
}
