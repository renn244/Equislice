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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	err := bootstrap.NewSentry(&cfg)
	if err != nil {
		fmt.Printf("Sentry initialization failed:  %v\n", err)
		return
	}

	traceProvider, err := bootstrap.InitTracer(ctx)
	if err != nil {
		fmt.Printf("Tracer initialization failed:  %v\n", err)
		return
	}

	meterProvider, err := bootstrap.InitMetrics(ctx)
	if err != nil {
		fmt.Printf("Metrics initialization failed:  %v\n", err)
		return
	}

	app, err := bootstrap.NewApp(&cfg)
	if err != nil {
		fmt.Printf("App initialization failed:  %v\n", err)
		return
	}

	server := &http.Server{
		Addr:    ":3000",
		Handler: app.Router,
	}

	go func() {
		log.Println("Server is Running on http://localhost:3000")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			sentry.CaptureException(err)
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Print("Shutdown signal received, shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server graceful shutdown failed: %v", err)
	}
	if err := traceProvider.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Tracer shutdown failed: %v", err)
	}
	if err := meterProvider.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Metrics shutdown failed: %v", err)
	}

	log.Print("Server shutdown complete")
}
