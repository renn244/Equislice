package services

import (
	"backend/internal/dto"
	"backend/internal/ports"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

type tracingDB struct {
	ports.DBService[dto.PanoramaEntity]
	deleted string
}

func (*tracingDB) Insert(_ context.Context, entity dto.PanoramaEntity, _ *string) (dto.PanoramaEntity, error) {
	return entity, nil
}
func (db *tracingDB) Delete(_ context.Context, id string) error { db.deleted = id; return nil }

type tracingQueue struct {
	ports.QueueService
	body        string
	spanContext trace.SpanContext
	err         error
}

func (queue *tracingQueue) Enqueue(ctx context.Context, body string) (string, error) {
	queue.body, queue.spanContext = body, trace.SpanContextFromContext(ctx)
	return body, queue.err
}

func TestPostPanoramaPropagatesProducerContext(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "sent", true: "send failure"}[fail], func(t *testing.T) {
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

			member, _ := baggage.NewMember("test", "panorama")
			bag, _ := baggage.New(member)
			ctx := baggage.ContextWithBaggage(context.Background(), bag)
			ctx, httpSpan := provider.Tracer("test").Start(ctx, "HTTP request", trace.WithSpanKind(trace.SpanKindServer))
			defer httpSpan.End()
			queue, db := &tracingQueue{}, &tracingDB{}
			if fail {
				queue.err = errors.New("queue unavailable")
			}
			service := NewPanoramaService(nil, nil, db, queue)
			id, err := service.PostPanorama(ctx, "source.jpg", 2, 3, "jpg", "tile_{row}_{col}")
			if (err != nil) != fail {
				t.Fatalf("error = %v", err)
			}
			var message dto.PanoramaQueueMessage
			if err := json.Unmarshal([]byte(queue.body), &message); err != nil {
				t.Fatal(err)
			}
			if message.JobId == "" || (!fail && message.JobId != id) {
				t.Fatalf("job = %q, result = %q", message.JobId, id)
			}
			extracted := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(message.TraceContext))
			parent := trace.SpanContextFromContext(extracted)
			if !parent.IsRemote() || !parent.Equal(queue.spanContext.WithRemote(true)) {
				t.Fatal("message does not carry the producer context")
			}
			if parent.TraceID() != httpSpan.SpanContext().TraceID() || parent.SpanID() == httpSpan.SpanContext().SpanID() {
				t.Fatal("producer did not continue the HTTP trace with a new span")
			}
			if baggage.FromContext(extracted).Member("test").Value() != "panorama" {
				t.Fatal("baggage was lost")
			}
			spans := recorder.Ended()
			if len(spans) != 1 || spans[0].SpanKind() != trace.SpanKindProducer || spans[0].Parent().SpanID() != httpSpan.SpanContext().SpanID() {
				t.Fatalf("producer spans = %v", spans)
			}
			if fail {
				if spans[0].Status().Code != codes.Error || len(spans[0].Events()) != 1 || db.deleted != message.JobId {
					t.Fatal("send failure was not traced or existing cleanup changed")
				}
			} else if spans[0].Status().Code == codes.Error || db.deleted != "" {
				t.Fatal("successful send reported failure or deleted the job")
			}
		})
	}
}
