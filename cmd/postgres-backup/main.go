package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/artarts36/postgres-backup/internal/application"
	"github.com/artarts36/postgres-backup/internal/config"
	"github.com/artarts36/postgres-backup/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to parse environment variables", slog.Any("err", err))
		os.Exit(1)
	}

	metricsCollector, err := createMetricsCollector(cfg.MetricsServer)
	if err != nil {
		slog.Error("failed to create metrics collector", slog.Any("err", err))
		os.Exit(1)
	}

	app, err := application.NewApp(cfg, metricsCollector)
	if err != nil {
		slog.Error("failed to create application", slog.Any("err", err))
		os.Exit(1)
	}

	defer app.Close()

	ctx := context.Background()

	err = app.Run(ctx)
	if ferr := metricsCollector.Flush(ctx); err != nil {
		slog.ErrorContext(ctx, "[main] failed to flush metrics", slog.Any("err", ferr))
	}

	if err != nil {
		slog.ErrorContext(ctx, "[main] failed to run application", slog.Any("err", err))
	}
}

func createMetricsCollector(
	metricsServer string,
) (metrics.Collector, error) {
	if metricsServer == "" {
		return metrics.NoopCollector{}, nil
	}

	registry := prometheus.NewRegistry()
	collector := metrics.NewPrometheusCollector("postgresbackup")

	if err := registry.Register(collectors.NewBuildInfoCollector()); err != nil {
		return nil, fmt.Errorf("register build info: %w", err)
	}

	if err := registry.Register(collector); err != nil {
		return nil, fmt.Errorf("register prometheus collector: %w", err)
	}

	return metrics.NewPushPrometheusCollector(collector, metricsServer, registry), nil
}
