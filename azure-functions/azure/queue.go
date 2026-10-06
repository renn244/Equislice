package azure

import (
	"context"
	"log/slog"

	"github.com/azure/azure-functions-golang-worker/sdk/bindings"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func ExtractQueueContext(ctx context.Context, carrier map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(carrier))
}

func LogQueueItem(ctx context.Context, msg bindings.QueueMessage) {
	slog.InfoContext(ctx, "Queue item found!",
		"id", msg.Id,
		"body", string(msg.Body),
		"dequeue_count", msg.DequeueCount,
		"pop_receipt", msg.PopReceipt,
		"expiration_time", msg.ExpirationTime,
		"insertion_time", msg.InsertionTime,
		"next_visible_time", msg.NextVisibleTime,
	)
}
