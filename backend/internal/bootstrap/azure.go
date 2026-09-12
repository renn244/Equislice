package bootstrap

import (
	"backend/internal/azure/blob"
	"backend/internal/azure/queue"
	"backend/internal/azure/table"
	"backend/internal/config"
	"backend/internal/util/constants"
	"fmt"

	"github.com/getsentry/sentry-go"
)

type AzureClients struct {
	PanoramaStorage      *blob.Client
	PanoramaSliceStorage *blob.Client
	PanoramaQueue        *queue.Client
	PanoramaTable        *table.Client
}

func NewAzure(cfg *config.Config) (*AzureClients, error) {
	panoramaStorage, err := blob.NewClient(blob.AzureBlobStorageConfig{
		ConnectionString: cfg.AzureConnectionString,
		Container:        constants.Container.Equirectangular,
	})
	if err != nil {
		sentry.CaptureException(err)
		fmt.Println(err)
		return nil, err
	}

	panoramaSliceStorage, err := blob.NewClient(blob.AzureBlobStorageConfig{
		ConnectionString: cfg.AzureConnectionString,
		Container:        constants.Container.EquirectangularSlice,
	})
	if err != nil {
		sentry.CaptureException(err)
		fmt.Println(err)
		return nil, err
	}

	panoramaQueue, err := queue.NewClient(queue.AzureQueueStorageConfig{
		ConnectionString: cfg.AzureConnectionString,
		Queue:            constants.Queue.PanoramaSlice,
	})
	if err != nil {
		sentry.CaptureException(err)
		fmt.Println(err)
		return nil, err
	}

	panoramaTable, err := table.NewClient(table.AzureTableDataConfig{
		ConnectionString: cfg.AzureConnectionString,
		Table:            constants.Table.Panorama,
	})
	if err != nil {
		sentry.CaptureException(err)
		fmt.Println(err)
		return nil, err
	}

	return &AzureClients{
		PanoramaStorage:      panoramaStorage,
		PanoramaSliceStorage: panoramaSliceStorage,
		PanoramaQueue:        panoramaQueue,
		PanoramaTable:        panoramaTable,
	}, nil
}
