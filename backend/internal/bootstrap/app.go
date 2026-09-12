package bootstrap

import (
	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/handler"
	"backend/internal/router"
	"backend/internal/services"
	"net/http"
)

type App struct {
	Router http.Handler
}

func NewApp(cfg *config.Config) (*App, error) {
	azure, err := NewAzure(cfg)
	if err != nil {
		return nil, err
	}

	// azure services
	panoramaTableService := services.NewTableStorageService[dto.PanoramaEntity](azure.PanoramaTable.Client, azure.PanoramaTable.AzureTableDataConfig.Table, "Panoramas")

	panoramaFileStorageService := services.NewFileStorageService(azure.PanoramaStorage.Client, azure.PanoramaStorage.AzureBlobStorageConfig.Container)
	panoramaSliceFileStorageService := services.NewFileStorageService(azure.PanoramaSliceStorage.Client, azure.PanoramaSliceStorage.AzureBlobStorageConfig.Container)

	panoramaQueueService := services.NewQueueService(azure.PanoramaQueue.Client, azure.PanoramaQueue.AzureQueueStorageConfig.Queue)

	// logic services and handler
	panoramaService := services.NewPanoramaService(panoramaFileStorageService, panoramaSliceFileStorageService, panoramaTableService, panoramaQueueService)
	panoramaHandler := handler.NewPanoramaHandler(panoramaService)

	r := router.NewRouter(cfg, panoramaHandler)

	return &App{
		Router: r,
	}, nil
}
