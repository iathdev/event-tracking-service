package signoz

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
)

func (s *SigNoz) HttpExporter(ctx context.Context) (log.Exporter, error) {
	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(s.OTLPAddress),
		otlploghttp.WithInsecure(),
	}

	if s.OTLPToken != "" {
		opts = append(opts, otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + s.OTLPToken,
		}))
	}

	return otlploghttp.New(ctx, opts...)
}
