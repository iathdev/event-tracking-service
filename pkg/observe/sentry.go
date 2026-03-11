package observe

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

const (
	sentryFlushTimeout = 2 * time.Second
)

type SentryConfig struct {
	Enable           bool
	DSN              string
	Environment      string
	Release          string
	SampleRate       float64
	TracesSampleRate float64
	Debug            bool
}

func InitSentry(cfg *SentryConfig) error {
	if !cfg.Enable {
		return nil
	}

	if cfg.DSN == "" {
		return fmt.Errorf("sentry DSN is required when sentry is enabled")
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		SampleRate:       cfg.SampleRate,
		TracesSampleRate: cfg.TracesSampleRate,
		Debug:            cfg.Debug,
		AttachStacktrace: true,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize sentry: %w", err)
	}

	return nil
}

func FlushSentry() {
	sentry.Flush(sentryFlushTimeout)
}

func CaptureException(err error) {
	sentry.CaptureException(err)
}

func CaptureMessage(message string) {
	sentry.CaptureMessage(message)
}

func RecoverWithSentry(err interface{}) {
	if err != nil {
		sentry.CurrentHub().Recover(err)
		sentry.Flush(sentryFlushTimeout)
	}
}
