package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/tracing/azotel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func tracingOptions() azcore.ClientOptions {
	return azcore.ClientOptions{
		TracingProvider: azotel.NewTracingProvider(otel.GetTracerProvider(), nil),
	}
}

// The current Blob SDK does not start operation spans, unlike the Table SDK.
func startBlobSpan(ctx context.Context, name string) (context.Context, func(error)) {
	ctx, span := otel.Tracer("azure-function/azure").Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient))
	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "blob operation failed")
		}
		span.End()
	}
}
