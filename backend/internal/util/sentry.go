package util

import (
	"context"

	"github.com/getsentry/sentry-go"
)

func CaptureSentryError(err error, ctx context.Context) {
	hub := sentry.GetHubFromContext(ctx)

	if hub != nil {
		hub.CaptureException(err)
	}
}
