package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"azure-function/bootstrap"

	"github.com/azure/azure-functions-golang-worker/sdk"
	"github.com/azure/azure-functions-golang-worker/worker"
	"github.com/getsentry/sentry-go"
)

func main() {
	app := sdk.FunctionApp()

	err := sentry.Init(sentry.ClientOptions{
		Dsn:         os.Getenv("SENTRY_DSN"),
		Environment: os.Getenv("SENTRY_ENVIRONMENT"),
	})
	if err != nil {
		slog.Error("Sentry initialization failed; continuing without Sentry", "error", err)
	} else {
		defer sentry.Flush(2 * time.Second)
	}

	traceProvider, err := bootstrap.InitTracer(context.Background())
	if err != nil {
		slog.Error("Tracer initialization failed; continuing without tracing", "error", err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := traceProvider.Shutdown(ctx); err != nil {
				slog.Error("Tracer shutdown failed", "error", err)
			}
		}()
	}

	app.Queue("processPanoramaSlice", processPanoramaSliceHandler,
		sdk.WithQueueName("panorama-slice"),
		sdk.WithConnection("AzureWebJobsStorage"),
	)

	app.Queue("processFailedPanoramaSliced", processFailedPanoramaSlicedHandler,
		sdk.WithQueueName("panorama-slice-poison"),
		sdk.WithConnection("AzureWebJobsStorage"),
	)

	worker.Start(app)
}
