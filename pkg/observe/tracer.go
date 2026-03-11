package observe

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const (
	tracerInitTimeout = 5 * time.Second
)

type TracingConfig struct {
	Enable      bool
	ServiceName string
	Endpoint    string
	UseGRPC     bool
	SampleRatio float64
	Environment string
}

func InitTracer(ctx context.Context, cfg *TracingConfig) (func(context.Context) error, error) {
	noopShutdown := func(context.Context) error { return nil }

	if !cfg.Enable {
		return noopShutdown, nil
	}

	// Create context with timeout for initialization
	initCtx, cancel := context.WithTimeout(ctx, tracerInitTimeout)
	defer cancel()

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return noopShutdown, err
	}

	var exporter *otlptrace.Exporter
	if cfg.UseGRPC {
		exporter, err = otlptrace.New(initCtx, otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithTimeout(tracerInitTimeout),
		))
	} else {
		exporter, err = otlptrace.New(initCtx, otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(cfg.Endpoint),
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithTimeout(tracerInitTimeout),
		))
	}
	if err != nil {
		return noopShutdown, err
	}

	var samplerOpt trace.TracerProviderOption
	if cfg.SampleRatio > 0 && cfg.SampleRatio < 1.0 {
		samplerOpt = trace.WithSampler(trace.TraceIDRatioBased(cfg.SampleRatio))
	} else {
		samplerOpt = trace.WithSampler(trace.AlwaysSample())
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		samplerOpt,
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return provider.Shutdown, nil
}
