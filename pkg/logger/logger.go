package logger

import (
	"event-tracking-service/pkg/logger/signoz"
	"context"
	"fmt"
	"os"
	"sync"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level           string // debug, info, warn, error
	Format          string // json, console
	ServiceName     string
	Channel         string // console, signoz
	OTLPEndpoint    string
	OTLPToken       string
	EnableGRPC      bool
	EnableAsync     bool
	AsyncBufferSize int
	BatchSize       int
	BatchTimeout    int
	Environment     string
}

var (
	signozConfigInstance *signoz.SigNoz
	signozConfigOnce     sync.Once
)

type LoggerFactory struct {
	config Config
}

func NewLoggerFactory(cfg Config) *LoggerFactory {
	return &LoggerFactory{config: cfg}
}

func (f *LoggerFactory) CreateLogger() (*zap.Logger, error) {
	var zapConfig zap.Config

	if f.config.Format == "console" {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	zapConfig.Level = getLogLevel(f.config.Level)

	if f.config.ServiceName != "" {
		zapConfig.InitialFields = map[string]interface{}{
			"service": f.config.ServiceName,
		}
	}

	return zapConfig.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}

func (f *LoggerFactory) getSignozConfig() *signoz.SigNoz {
	signozConfigOnce.Do(func() {
		signozConfigInstance = &signoz.SigNoz{
			Service:         f.config.ServiceName,
			OTLPAddress:     f.config.OTLPEndpoint,
			OTLPToken:       f.config.OTLPToken,
			EnableGRPC:      f.config.EnableGRPC,
			EnableAsync:     f.config.EnableAsync,
			AsyncBufferSize: f.config.AsyncBufferSize,
			BatchSize:       f.config.BatchSize,
			BatchTimeout:    f.config.BatchTimeout,
		}
	})
	return signozConfigInstance
}

func (f *LoggerFactory) CreateSigNozLogger(ctx context.Context) (*zap.Logger, *log.LoggerProvider, error) {
	if f.config.OTLPEndpoint == "" {
		return nil, nil, fmt.Errorf("OTLP endpoint is required for SigNoz logging")
	}

	signozConfig := f.getSignozConfig()

	exporter := signozConfig.Exporter(ctx)
	if exporter == nil {
		return nil, nil, fmt.Errorf("failed to create SigNoz exporter")
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(f.config.ServiceName),
		attribute.String("environment", f.config.Environment),
	)

	provider := signozConfig.Provider(exporter, res)

	core := otelzap.NewCore(
		f.config.ServiceName,
		otelzap.WithLoggerProvider(provider),
	)

	logLevel := getLogLevel(f.config.Level)

	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	).WithOptions(zap.IncreaseLevel(logLevel))

	return logger, provider, nil
}

// NewLogger creates a logger based on config (backward compatible)
func NewLogger(cfg *Config) (*zap.Logger, error) {
	factory := NewLoggerFactory(*cfg)
	return factory.CreateLogger()
}

func getLogLevel(level string) zap.AtomicLevel {
	switch level {
	case "debug":
		return zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
}

func WithContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	if ctx == nil {
		return logger
	}

	// Get trace ID from OpenTelemetry span
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return logger.With(
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	// Get trace ID from context value
	if traceID, ok := ctx.Value("trace_id").(string); ok && traceID != "" {
		return logger.With(zap.String("trace_id", traceID))
	}

	return logger
}

func NewNop() *zap.Logger {
	return zap.NewNop()
}

func Must(logger *zap.Logger, err error) *zap.Logger {
	if err != nil {
		os.Exit(1)
	}
	return logger
}
