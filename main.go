package main

import (
	"context"
	"event-tracking-service/config"
	"event-tracking-service/internal/httpserver"
	"event-tracking-service/internal/repositories"
	"event-tracking-service/internal/scheduler"
	"event-tracking-service/internal/services"
	"event-tracking-service/pkg/database"
	"event-tracking-service/pkg/observe"
	"log"

	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()
	ctx := context.Background()

	// Initialize all observability (logger, tracing, sentry)
	obs, err := observe.Init(ctx, cfg.GetObservabilityConfig())
	if err != nil {
		log.Fatalf("Failed to initialize observability: %v", err)
	}
	defer obs.Shutdown(ctx)

	zapLogger := obs.Logger

	// Initialize database connections
	dbConns, err := database.NewConnections(cfg)
	if err != nil {
		zapLogger.Fatal("Could not initialize database connections", zap.Error(err))
	}
	defer dbConns.Close()

	// Enable GORM tracing if tracing is enabled
	if obs.TracingEnabled() {
		if err := observe.WithGormTracing(dbConns.DB, "postgresql"); err != nil {
			zapLogger.Warn("Failed to enable GORM tracing", zap.Error(err))
		} else {
			zapLogger.Info("GORM tracing enabled")
		}
	}

	// Initialize services
	eventBuffer := services.NewEventBuffer(dbConns.Redis, cfg, zapLogger)
	trackingEventRepo := repositories.NewTrackingEventRepository(dbConns.DB)
	eventProcessor := services.NewEventProcessor(eventBuffer, trackingEventRepo, cfg, zapLogger)

	// Initialize scheduler
	sched, err := scheduler.NewScheduler(dbConns.Redis, cfg, zapLogger, eventProcessor)
	if err != nil {
		zapLogger.Fatal("Failed to create scheduler", zap.Error(err))
	}

	if err := sched.RegisterJobs(); err != nil {
		zapLogger.Fatal("Failed to register scheduler jobs", zap.Error(err))
	}

	sched.Start()
	defer func() {
		if err := sched.Stop(); err != nil {
			zapLogger.Error("Failed to stop scheduler", zap.Error(err))
		}
	}()

	zapLogger.Info("Starting HTTP server", zap.String("port", cfg.Server.Port))

	if err := httpserver.StartHTTPServer(cfg, zapLogger, eventBuffer); err != nil {
		zapLogger.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}
