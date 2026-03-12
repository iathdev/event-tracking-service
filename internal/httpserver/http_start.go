package httpserver

import (
	"context"
	"errors"
	"event-tracking-service/config"
	"event-tracking-service/internal/handlers"
	"event-tracking-service/internal/middleware"
	"event-tracking-service/internal/services"
	"event-tracking-service/pkg/common"
	"event-tracking-service/pkg/database"
	"event-tracking-service/pkg/observe"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	shutdownTimeout = 10 * time.Second
)

func StartHTTPServer(cfg *config.Config, logger *zap.Logger, eventBuffer *services.EventBuffer) error {
	configureGinMode(cfg.Server.Env)

	conn, err := database.NewConnections(cfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	server := New(cfg)
	registerMiddleware(server.Engine(), cfg, logger)
	registerRoutes(server.Engine(), conn, cfg, logger, eventBuffer)

	serverErrors := make(chan error, 1)
	go launchServerAsync(server, serverErrors)

	return runServerLifecycle(server, serverErrors, logger)
}

func configureGinMode(env string) {
	if env == "local" || env == "develop" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

func launchServerAsync(server *Server, errChan chan<- error) {
	if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errChan <- err
	}
}

func runServerLifecycle(server *Server, serverErrors chan error, logger *zap.Logger) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return err
	case <-quit:
		return executeGracefulRunShutdown(server, logger)
	}
}

func executeGracefulRunShutdown(server *Server, logger *zap.Logger) error {
	logger.Info("Initiating server shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return errors.New("server shutdown failed: " + err.Error())
	}

	logger.Info("Server shutdown completed")
	return nil
}

func registerMiddleware(router *gin.Engine, cfg *config.Config, logger *zap.Logger) {
	router.Use(gin.Recovery())
	router.Use(observe.RecoveryMiddleware(logger))

	if cfg.Sentry.Enable {
		router.Use(sentrygin.New(sentrygin.Options{
			Repanic: true,
		}))
	}

	if cfg.Tracing.Enable {
		router.Use(observe.GinMiddleware(cfg.Tracing.ServiceName))
	}

	router.Use(observe.LoggingMiddleware(logger))
}

func registerRoutes(router *gin.Engine, conn *database.Connections, cfg *config.Config, logger *zap.Logger, eventBuffer *services.EventBuffer) {
	healthCheckHandler := handlers.NewHealthCheckHandler()
	docsHandler := handlers.NewDocsHandler()
	trackingEventHandler := handlers.NewTrackingEventHandler(eventBuffer, logger, cfg.EventBuffer.SensitiveFields)
	monitoringHandler := handlers.NewMonitoringHandler(eventBuffer, logger)

	router.GET("/health", healthCheckHandler.Health)

	// API Documentation
	router.GET("/docs", docsHandler.ScalarUI)
	router.GET("/docs/openapi.yaml", docsHandler.OpenAPISpec)

	api := router.Group("/api/v1/")
	{
		events := api.Group("/events")
		{
			events.POST("", trackingEventHandler.Create)
			events.POST("/batch", trackingEventHandler.CreateBatch)
		}

		monitoring := api.Group("/monitoring")
		monitoring.Use(middleware.RequiredApiKeyInternal())
		{
			monitoring.GET("/queue-stats", monitoringHandler.QueueStats)
		}
	}

	router.NoRoute(func(c *gin.Context) {
		common.WriteErrorResponseWithCode(c, "Not Found", nil, http.StatusNotFound)
	})
}
