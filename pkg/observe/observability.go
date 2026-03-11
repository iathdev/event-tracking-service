package observe

import (
	"event-tracking-service/config"
	"event-tracking-service/pkg/logger"
	"context"
	"log"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
)

// Observability manages all observability concerns (logging, tracing, error tracking)
type Observability struct {
	Logger         *zap.Logger
	logProvider    *sdklog.LoggerProvider
	shutdowns      []func() error
	tracingEnabled bool
}

// Init initializes all observability components and returns an Observability instance
func Init(ctx context.Context, cfg *config.ObservabilityConfig) (*Observability, error) {
	obs := &Observability{
		shutdowns: make([]func() error, 0),
	}

	// Initialize logger
	if err := obs.initLogger(ctx, cfg); err != nil {
		return nil, err
	}

	// Initialize tracing (non-fatal)
	obs.initTracing(ctx, cfg)

	// Initialize Sentry (non-fatal)
	obs.initSentry(cfg)

	return obs, nil
}

func (o *Observability) initLogger(ctx context.Context, cfg *config.ObservabilityConfig) error {
	loggerFactory := logger.NewLoggerFactory(logger.Config{
		Level:           cfg.LogLevel,
		Format:          cfg.LogFormat,
		ServiceName:     cfg.ServiceName,
		Channel:         cfg.LogChannel,
		OTLPEndpoint:    cfg.OTLPEndpoint,
		OTLPToken:       cfg.OTLPToken,
		EnableGRPC:      cfg.EnableGRPC,
		EnableAsync:     cfg.EnableAsync,
		AsyncBufferSize: cfg.AsyncBufferSize,
		BatchSize:       cfg.BatchSize,
		BatchTimeout:    cfg.BatchTimeout,
		Environment:     cfg.Environment,
	})

	if cfg.LogChannel == "signoz" {
		zapLogger, logProvider, err := loggerFactory.CreateSigNozLogger(ctx)
		if err != nil {
			log.Printf("Warning: Failed to initialize SigNoz logger: %v. Falling back to console logger.", err)
			zapLogger, err = loggerFactory.CreateLogger()
			if err != nil {
				return err
			}
			o.Logger = zapLogger
		} else {
			o.Logger = zapLogger
			o.logProvider = logProvider
			log.Println("SigNoz logger initialized")
		}
	} else {
		zapLogger, err := loggerFactory.CreateLogger()
		if err != nil {
			return err
		}
		o.Logger = zapLogger
	}

	// Add logger sync to shutdowns
	o.shutdowns = append(o.shutdowns, func() error {
		return o.Logger.Sync()
	})

	return nil
}

func (o *Observability) initTracing(ctx context.Context, cfg *config.ObservabilityConfig) {
	shutdownTracer, err := InitTracer(ctx, &TracingConfig{
		Enable:      cfg.TracingEnable,
		ServiceName: cfg.ServiceName,
		Endpoint:    cfg.TracingEndpoint,
		UseGRPC:     cfg.TracingUseGRPC,
		SampleRatio: cfg.TracingSampleRatio,
		Environment: cfg.Environment,
	})

	if err != nil {
		o.Logger.Warn("Failed to initialize tracer, continuing without tracing", zap.Error(err))
		return
	}

	if cfg.TracingEnable {
		o.tracingEnabled = true
		o.Logger.Info("Tracing enabled",
			zap.String("endpoint", cfg.TracingEndpoint),
			zap.Bool("grpc", cfg.TracingUseGRPC),
		)
		o.shutdowns = append(o.shutdowns, func() error {
			return shutdownTracer(context.Background())
		})
	}
}

// TracingEnabled returns whether tracing is enabled
func (o *Observability) TracingEnabled() bool {
	return o.tracingEnabled
}

func (o *Observability) initSentry(cfg *config.ObservabilityConfig) {
	if err := InitSentry(&SentryConfig{
		Enable:           cfg.SentryEnable,
		DSN:              cfg.SentryDSN,
		Environment:      cfg.Environment,
		SampleRate:       cfg.SentrySampleRate,
		TracesSampleRate: cfg.SentryTracesSampleRate,
		Debug:            cfg.SentryDebug,
	}); err != nil {
		o.Logger.Warn("Failed to initialize Sentry, continuing without error tracking", zap.Error(err))
		return
	}

	if cfg.SentryEnable {
		o.Logger.Info("Sentry enabled", zap.String("environment", cfg.Environment))
		o.shutdowns = append(o.shutdowns, func() error {
			FlushSentry()
			return nil
		})
	}
}

// Shutdown gracefully shuts down all observability components
func (o *Observability) Shutdown(ctx context.Context) {
	// Shutdown log provider first (if using SigNoz)
	if o.logProvider != nil {
		if err := o.logProvider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down log provider: %v", err)
		}
	}

	// Shutdown in reverse order (LIFO)
	for i := len(o.shutdowns) - 1; i >= 0; i-- {
		if err := o.shutdowns[i](); err != nil {
			log.Printf("Error during observability shutdown: %v", err)
		}
	}
}
