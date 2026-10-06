package bootstrap

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/proto"
)

func TestTracerShutdownExportsConfiguredResource(t *testing.T) {
	requests := make(chan *collectortrace.ExportTraceServiceRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/traces" || r.Header.Get("X-Test") != "configured" {
			t.Errorf("unexpected OTLP request path or header")
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		request := &collectortrace.ExportTraceServiceRequest{}
		if err := proto.Unmarshal(data, request); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		requests <- request
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", server.URL+"/v1/traces")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "X-Test=configured")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_HEADERS", "X-Test=configured")
	t.Setenv("OTEL_EXPORTER_OTLP_COMPRESSION", "none")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_COMPRESSION", "none")
	t.Setenv("OTEL_SERVICE_NAME", "equislice-function")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "deployment.environment=testing")
	previousProvider, previousPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	t.Cleanup(func() { otel.SetTracerProvider(previousProvider); otel.SetTextMapPropagator(previousPropagator) })
	provider, err := InitTracer(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	_, span := otel.Tracer("test").Start(context.Background(), "queued panorama")
	span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := provider.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case request := <-requests:
		if len(request.ResourceSpans) != 1 {
			t.Fatalf("resource spans = %d", len(request.ResourceSpans))
		}
		resource := request.ResourceSpans[0]
		attributes := map[string]string{}
		for _, attr := range resource.Resource.Attributes {
			attributes[attr.Key] = attr.Value.GetStringValue()
		}
		if attributes["service.name"] != "equislice-function" || attributes["deployment.environment"] != "testing" {
			t.Fatalf("resource = %v", attributes)
		}
		if len(resource.ScopeSpans) != 1 || len(resource.ScopeSpans[0].Spans) != 1 || resource.ScopeSpans[0].Spans[0].Name != "queued panorama" {
			t.Fatal("shutdown did not flush the processing span")
		}
	default:
		t.Fatal("shutdown did not export traces")
	}
}
