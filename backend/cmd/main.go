package main

import (
	"backend/internal/azure/blob"
	"backend/internal/azure/queue"
	"backend/internal/azure/table"
	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/handler"
	"backend/internal/services"
	"backend/internal/util/constants"
	"encoding/json"
	"fmt"

	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	cfg := config.Load()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendUrl},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
	}))

	// azure instances
	panoramaStorage, err := blob.NewClient(blob.AzureBlobStorageConfig{
		ConnectionString: cfg.AzureConnectionString,
		Container:        constants.Container.Equirectangular,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	panoramaSliceStorage, err := blob.NewClient(blob.AzureBlobStorageConfig{
		ConnectionString: cfg.AzureConnectionString,
		Container:        constants.Container.EquirectangularSlice,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	panoramaQueue, err := queue.NewClient(queue.AzureQueueStorageConfig{
		ConnectionString: cfg.AzureConnectionString,
		Queue:            constants.Queue.PanoramaSlice,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	panoramaTable, err := table.NewClient(table.AzureTableDataConfig{
		ConnectionString: cfg.AzureConnectionString,
		Table:            constants.Table.Panorama,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// azure services
	panoramaTableService := services.NewTableStorageService[dto.PanoramaEntity](panoramaTable.Client, panoramaTable.AzureTableDataConfig.Table, "Panoramas")

	panoramaFileStorageService := services.NewFileStorageService(panoramaStorage.Client, panoramaStorage.AzureBlobStorageConfig.Container)
	panoramaSliceFileStorageService := services.NewFileStorageService(panoramaSliceStorage.Client, panoramaSliceStorage.AzureBlobStorageConfig.Container)

	panoramaQueueService := services.NewQueueService(panoramaQueue.Client, panoramaQueue.AzureQueueStorageConfig.Queue)

	// logic services and handler
	panoramaService := services.NewPanoramaService(panoramaFileStorageService, panoramaSliceFileStorageService, panoramaTableService, panoramaQueueService)
	panoramaHandler := handler.NewPanoramaHandler(panoramaService)

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

	log.Println("Server is Running on http://localhost:3000")
	err = http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}
