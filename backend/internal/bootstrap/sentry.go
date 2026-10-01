package bootstrap

import (
	"backend/internal/config"

	"github.com/getsentry/sentry-go"
)

func NewSentry(cfg *config.Config) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.SentryEnvironment,
		EnableTracing:    true,
		TracesSampleRate: 0.2,
		SendDefaultPII:   false,
	})
}
