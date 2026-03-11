package signoz

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

var (
	DefaultBatchSize     = 50
	DefaultQueueSize     = 2048
	DefaultExportTimeout = 10
)

type SigNoz struct {
	Service         string
	OTLPAddress     string
	OTLPToken       string
	EnableGRPC      bool
	EnableAsync     bool
	AsyncBufferSize int
	BatchSize       int
	BatchTimeout    int

	exporterMu  sync.RWMutex
	exporter    log.Exporter

	processorMu sync.RWMutex
	processor   log.Processor

	providerMu  sync.RWMutex
	provider    *log.LoggerProvider
}

func (s *SigNoz) GetService() string {
	if s.Service == "" {
		return "unknown"
	}
	return s.Service
}

func (s *SigNoz) GetBatchSize() int {
	if s.BatchSize <= 0 {
		return DefaultBatchSize
	}
	return s.BatchSize
}

func (s *SigNoz) GetQueueSize() int {
	if s.AsyncBufferSize <= 0 {
		return DefaultQueueSize
	}
	return s.AsyncBufferSize
}

func (s *SigNoz) GetExportTimeout() time.Duration {
	if s.BatchTimeout <= 0 {
		return time.Duration(DefaultExportTimeout) * time.Second
	}
	return time.Duration(s.BatchTimeout) * time.Second
}

func (s *SigNoz) Processor(exporter log.Exporter) log.Processor {
	s.processorMu.RLock()
	if s.processor != nil {
		defer s.processorMu.RUnlock()
		return s.processor
	}
	s.processorMu.RUnlock()

	s.processorMu.Lock()
	defer s.processorMu.Unlock()

	if s.processor != nil {
		return s.processor
	}

	if s.EnableAsync {
		s.processor = log.NewBatchProcessor(
			exporter,
			log.WithExportMaxBatchSize(s.GetBatchSize()),
			log.WithExportTimeout(s.GetExportTimeout()),
			log.WithMaxQueueSize(s.GetQueueSize()),
		)
	} else {
		s.processor = log.NewSimpleProcessor(exporter)
	}

	return s.processor
}

func (s *SigNoz) Provider(exporter log.Exporter, res *resource.Resource) *log.LoggerProvider {
	s.providerMu.RLock()
	if s.provider != nil {
		defer s.providerMu.RUnlock()
		return s.provider
	}
	s.providerMu.RUnlock()

	s.providerMu.Lock()
	defer s.providerMu.Unlock()

	if s.provider != nil {
		return s.provider
	}

	s.provider = log.NewLoggerProvider(
		log.WithProcessor(s.Processor(exporter)),
		log.WithResource(res),
	)

	return s.provider
}

func (s *SigNoz) Exporter(ctx context.Context) log.Exporter {
	s.exporterMu.RLock()
	if s.exporter != nil {
		defer s.exporterMu.RUnlock()
		return s.exporter
	}
	s.exporterMu.RUnlock()

	s.exporterMu.Lock()
	defer s.exporterMu.Unlock()

	if s.exporter != nil {
		return s.exporter
	}

	var err error
	if s.EnableGRPC {
		s.exporter, err = s.GRPCExporter(ctx)
	} else {
		s.exporter, err = s.HttpExporter(ctx)
	}
	if err != nil {
		fmt.Printf("[SIGNOZ-LOG] Error creating OTLP exporter: %s\n", err.Error())
	}

	return s.exporter
}

func (s *SigNoz) Close(ctx context.Context) error {
	s.providerMu.Lock()
	defer s.providerMu.Unlock()

	if s.provider != nil {
		err := s.provider.Shutdown(ctx)
		s.provider = nil

		if err != nil {
			fmt.Printf("[SIGNOZ-LOG] Error shutting down logger provider: %s\n", err.Error())
		}

		s.processorMu.Lock()
		s.processor = nil
		s.processorMu.Unlock()

		s.exporterMu.Lock()
		s.exporter = nil
		s.exporterMu.Unlock()

		return err
	}
	return nil
}
