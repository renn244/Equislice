package bootstrap

import (
	"backend/internal/config"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

func NewSentry(cfg *config.Config) error {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.SentryEnvironment,
		EnableTracing:    true,
		TracesSampleRate: 0.2,
		SendDefaultPII:   false,
	}); err != nil {
		fmt.Printf("Sentry initialization failed:  %v\n", err)
		return err
	}
	defer sentry.Flush(2 * time.Second)

	return nil
}
