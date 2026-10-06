package azure

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestStorageOperationsAreChildrenOfProcessingSpan(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { otel.SetTracerProvider(previous); _ = provider.Shutdown(context.Background()) })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"job_id":"job","row":2,"column":3}`))
		} else if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
		} else {
			t.Errorf("unexpected request method: %s", r.Method)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()
	connection := "DefaultEndpointsProtocol=http;AccountName=test;AccountKey=ZmFrZS1rZXk=;BlobEndpoint=" + server.URL + ";TableEndpoint=" + server.URL
	table, err := CreateTableInstance("Panorama", connection)
	if err != nil {
		t.Fatal(err)
	}
	blobs, err := CreateBlobInstance(connection)
	if err != nil {
		t.Fatal(err)
	}
	ctx, processing := provider.Tracer("test").Start(context.Background(), "process panorama")
	defer processing.End()
	if _, err := GetJobEntity(ctx, table, "job"); err != nil {
		t.Fatal(err)
	}
	if _, err := GetPropertiesBlob(ctx, blobs, "equirectangular-slice", "job\\0_0.jpg"); err != nil {
		t.Fatal(err)
	}
	spans := recorder.Ended()
	if len(spans) != 3 {
		t.Fatalf("storage spans = %d", len(spans))
	}
	parents := map[string]string{}
	ids := map[string]string{}
	for _, span := range spans {
		if span.SpanContext().TraceID() != processing.SpanContext().TraceID() {
			t.Fatalf("storage span %s lost processing trace", span.Name())
		}
		parents[span.Name()] = span.Parent().SpanID().String()
		ids[span.Name()] = span.SpanContext().SpanID().String()
	}
	if parents["Client.GetEntity"] != processing.SpanContext().SpanID().String() || parents["Blob.GetProperties"] != processing.SpanContext().SpanID().String() || parents["HTTP GET"] != ids["Client.GetEntity"] {
		t.Fatalf("unexpected storage trace parents: %v", parents)
	}
}
