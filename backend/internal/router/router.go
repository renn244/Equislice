package router

import (
	"backend/internal/config"
	"backend/internal/handler"
	"encoding/json"
	"log"
	"net/http"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewRouter(cfg *config.Config, panoramaHandler *handler.PanoramaHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(
		otelhttp.NewMiddleware(
			"equislice-http",
			otelhttp.WithSpanNameFormatter(func(operationName string, r *http.Request) string {
				if chi.RouteContext(r.Context()).RoutePattern() != "" {
					return r.Method + " " + chi.RouteContext(r.Context()).RoutePattern()
				}

				return r.Method + "<No-Pattern>"
			}),
		),
	)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	}).Handle)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendUrl},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
	}))

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"health": "ok",
		})

		log.Println("Server is Healthy.")
	})

	r.Post("/api/panorama/slice", panoramaHandler.PostPanorama)
	r.Post("/api/panorama/upload", panoramaHandler.GetUploadUrl)
	r.Get("/api/panorama/status", panoramaHandler.GetStatus)
	r.Get("/api/panorama/download", panoramaHandler.GetShareUrl)
	r.Get("/api/panorama/download-all", panoramaHandler.GetArchive)

	return r
}
