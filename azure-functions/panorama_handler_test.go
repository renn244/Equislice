package main

import (
	"context"
	"encoding/json"
	"testing"

	"azure-function/azure"

	"github.com/azure/azure-functions-golang-worker/sdk/bindings"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func tracingFixture(t *testing.T) (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	t.Helper()
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previousProvider, previousPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
		_ = provider.Shutdown(context.Background())
	})
	return provider, recorder
}

var producerCarrier = map[string]string{
	"traceparent": "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01",
	"tracestate":  "vendor=value",
	"baggage":     "test=panorama",
}

func TestHandlersContinueProducerTraceAndRecordErrors(t *testing.T) {
	_, recorder := tracingFixture(t)
	// Fail before accessing Azure so the actual handler's error path is exercised.
	t.Setenv("EQUISLICE_STORAGE_CONNECTION_STRING", "invalid")
	payload, err := json.Marshal(azure.PanoramaSliceBody{JobId: "job", TraceContext: producerCarrier})
	if err != nil {
		t.Fatal(err)
	}
	message := bindings.QueueMessage{Id: "message", Body: payload, DequeueCount: 1}
	if err := processPanoramaSliceHandler(context.Background(), message); err == nil {
		t.Fatal("expected existing storage initialization error")
	}
	message.DequeueCount = 2
	if err := processPanoramaSliceHandler(context.Background(), message); err == nil {
		t.Fatal("expected retry storage error")
	}
	if err := processFailedPanoramaSlicedHandler(context.Background(), message); err == nil {
		t.Fatal("expected poison storage error")
	}
	parent := trace.SpanContextFromContext(azure.ExtractQueueContext(context.Background(), producerCarrier))
	spans := recorder.Ended()
	if len(spans) != 3 {
		t.Fatalf("spans = %d", len(spans))
	}
	ids := map[trace.SpanID]bool{}
	for _, span := range spans {
		if span.SpanContext().TraceID() != parent.TraceID() || span.Parent().SpanID() != parent.SpanID() || !span.Parent().IsRemote() {
			t.Fatal("consumer lost its remote producer parent")
		}
		if span.SpanContext().TraceState().String() != "vendor=value" {
			t.Fatal("tracestate was lost")
		}
		if span.SpanKind() != trace.SpanKindConsumer || span.Status().Code != codes.Error || len(span.Events()) != 1 {
			t.Fatal("consumer failure was not recorded")
		}
		if ids[span.SpanContext().SpanID()] {
			t.Fatal("retry reused an ended span")
		}
		ids[span.SpanContext().SpanID()] = true
	}
	if spans[0].Name() != "panorama-slice process" || spans[2].Name() != "panorama-slice-poison process" {
		t.Fatal("handlers used incorrect queue span names")
	}
}

func TestQueueTracePreservesInvocationContext(t *testing.T) {
	provider, recorder := tracingFixture(t)
	type contextKey struct{}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "invocation"))
	defer cancel()
	ctx, invocation := provider.Tracer("test").Start(ctx, "host invocation")
	defer invocation.End()
	ctx, endSpan := startPanoramaQueueSpan(ctx, bindings.QueueMessage{Id: "message"}, azure.PanoramaSliceBody{JobId: "job", TraceContext: producerCarrier}, "panorama-slice")
	if ctx.Value(contextKey{}) != "invocation" || baggage.FromContext(ctx).Member("test").Value() != "panorama" {
		t.Fatal("context values or baggage were lost")
	}
	cancel()
	if ctx.Err() != context.Canceled {
		t.Fatal("invocation cancellation was lost")
	}
	endSpan(nil)
	links := recorder.Ended()[0].Links()
	if len(links) != 2 || links[1].SpanContext.SpanID() != invocation.SpanContext().SpanID() {
		t.Fatal("host invocation correlation was lost")
	}
}

func TestLegacyAndInvalidTraceContextStartNewTraces(t *testing.T) {
	_, recorder := tracingFixture(t)
	for _, carrier := range []map[string]string{nil, {"traceparent": "invalid"}} {
		ctx, endSpan := startPanoramaQueueSpan(context.Background(), bindings.QueueMessage{}, azure.PanoramaSliceBody{JobId: "job", TraceContext: carrier}, "panorama-slice")
		if !trace.SpanContextFromContext(ctx).IsValid() {
			t.Fatal("legacy message did not get a new span")
		}
		endSpan(nil)
	}
	for _, span := range recorder.Ended() {
		if span.Parent().IsValid() {
			t.Fatal("invalid or missing trace context became a parent")
		}
	}
}
