package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/config"
	"backend/internal/handler"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// newTestRouter installs recording trace and metric providers. They must be set
// before NewRouter, because otelhttp picks up the global providers when the
// middleware is created.
func newTestRouter(t *testing.T) (http.Handler, *tracetest.SpanRecorder, *sdkmetric.ManualReader) {
	t.Helper()

	spans := tracetest.NewSpanRecorder()
	metrics := sdkmetric.NewManualReader()

	previousTracer, previousMeter := otel.GetTracerProvider(), otel.GetMeterProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans)))
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(metrics)))
	t.Cleanup(func() {
		otel.SetTracerProvider(previousTracer)
		otel.SetMeterProvider(previousMeter)
	})

	r := NewRouter(&config.Config{FrontendUrl: "http://localhost:5173"}, &handler.PanoramaHandler{})
	return r, spans, metrics
}

func routeAttr(attrs []attribute.KeyValue) string {
	for _, attr := range attrs {
		if attr.Key == "http.route" {
			return attr.Value.AsString()
		}
	}
	return ""
}

func TestSpanNamesUseMatchedRoute(t *testing.T) {
	r, recorder, _ := newTestRouter(t)

	tests := []struct {
		path      string
		wantName  string
		wantRoute string
	}{
		{path: "/api/health", wantName: "GET /api/health", wantRoute: "/api/health"},
		{path: "/does-not-exist", wantName: "GET<No-Pattern>"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			recorder.Reset()
			r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, tt.path, nil))

			spans := recorder.Ended()
			if len(spans) != 1 {
				t.Fatalf("got %d spans, want 1", len(spans))
			}
			if got := spans[0].Name(); got != tt.wantName {
				t.Errorf("span name = %q, want %q", got, tt.wantName)
			}
			if got := routeAttr(spans[0].Attributes()); got != tt.wantRoute {
				t.Errorf("span http.route = %q, want %q", got, tt.wantRoute)
			}
		})
	}
}

func TestRequestDurationMetricHasRoute(t *testing.T) {
	r, _, reader := newTestRouter(t)

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/health", nil))

	var data metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &data); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}

	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != "http.server.request.duration" {
				continue
			}
			hist, ok := m.Data.(metricdata.Histogram[float64])
			if !ok || len(hist.DataPoints) == 0 {
				t.Fatalf("unexpected data for %s: %T", m.Name, m.Data)
			}
			if got := routeAttr(hist.DataPoints[0].Attributes.ToSlice()); got != "/api/health" {
				t.Fatalf("metric http.route = %q, want %q", got, "/api/health")
			}
			return
		}
	}
	t.Fatal("http.server.request.duration was not recorded")
}
