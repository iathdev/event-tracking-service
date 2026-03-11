package signoz

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *SigNoz) GRPCExporter(ctx context.Context) (log.Exporter, error) {
	conn, err := grpc.NewClient(
		s.OTLPAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	opts := []otlploggrpc.Option{
		otlploggrpc.WithGRPCConn(conn),
	}

	if s.OTLPToken != "" {
		opts = append(opts, otlploggrpc.WithHeaders(map[string]string{
			"Authorization": "Bearer " + s.OTLPToken,
		}))
	}

	return otlploggrpc.New(ctx, opts...)
}
