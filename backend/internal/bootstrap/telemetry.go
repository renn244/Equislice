package bootstrap

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	// initialize the resource which is what will contain the actual traces telemetry. with from env it will contain things like production and app name and more
	// with telemetry sdk it will give the telemetry sdk version when sending the telemetry
	res, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	// this is the exporter which will export it in a conventional http requests. we don't configure where it will send it yet.
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	// this is the provider which actually is the actual provider that will be use to handle the traces and batching and stuff like that its like the manager.
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)

	// we just register it in the otel so its becomes global
	otel.SetTracerProvider(traceProvider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return traceProvider, nil
}

func InitMetrics(ctx context.Context) (*sdkmetric.MeterProvider, error) {
	res, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHostID(),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	exporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create metric exporter: %w", err)
	}

	reader := sdkmetric.NewPeriodicReader(exporter)

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)

	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}
