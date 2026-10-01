package main

import (
	"backend/internal/bootstrap"
	"backend/internal/config"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
)

func main() {
	if err := run(); err != nil {
		log.Printf("Backend stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	if err := bootstrap.NewSentry(&cfg); err != nil {
		log.Printf("Sentry initialization failed; continuing without Sentry: %v", err)
	} else {
		defer sentry.Flush(2 * time.Second)
	}

	traceProvider, err := bootstrap.InitTracer(ctx)
	if err != nil {
		log.Printf("Tracer initialization failed; continuing without tracing: %v", err)
	} else {
		defer shutdownProvider("tracer", traceProvider.Shutdown)
	}

	meterProvider, err := bootstrap.InitMetrics(ctx)
	if err != nil {
		log.Printf("Metrics initialization failed; continuing without metrics: %v", err)
	} else {
		defer shutdownProvider("metrics", meterProvider.Shutdown)
	}

	app, err := bootstrap.NewApp(&cfg)
	if err != nil {
		return fmt.Errorf("initialize application: %w", err)
	}

	server := &http.Server{
		Addr:    ":3000",
		Handler: app.Router,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Println("Server is Running on http://localhost:3000")
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Print("Shutdown signal received, shutting down gracefully")
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			sentry.CaptureException(err)
			return fmt.Errorf("serve HTTP requests: %w", err)
		}
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("gracefully shut down server: %w", err)
	}

	log.Print("Server shutdown complete")
	return nil
}

func shutdownProvider(name string, shutdown func(context.Context) error) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := shutdown(shutdownCtx); err != nil {
		log.Printf("%s shutdown failed: %v", name, err)
	}
}
