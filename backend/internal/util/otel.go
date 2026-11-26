package util

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

// RouteTelemetry adds the matched chi route to the request's span and HTTP
// metrics. Register it after the otelhttp middleware: chi only knows the route
// once it has routed the request, so this reads it after next returns, while
// otelhttp's span and labeler are still open.
func RouteTelemetry(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		pattern := chi.RouteContext(r.Context()).RoutePattern()
		if pattern == "" {
			return
		}
		route := semconv.HTTPRoute(pattern)

		span := trace.SpanFromContext(r.Context())
		span.SetName(r.Method + " " + pattern)
		span.SetAttributes(route)

		// otelhttp records its metrics after this middleware returns, using these labels.
		if labeler, ok := otelhttp.LabelerFromContext(r.Context()); ok {
			labeler.Add(route)
		}
	})
}
