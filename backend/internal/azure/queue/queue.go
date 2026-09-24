package queue

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue/v2"
)

var (
	ErrInvalidConnectionString = errors.New("invalid connection string")
	ErrQueueNotFound           = errors.New("queue not found")
	ErrMessageNotFound         = errors.New("message not found")
	ErrInvalidInput            = errors.New("invalid input parameters")
)

type AzureQueueStorageConfig struct {
	Queue            string
	ConnectionString string
	Options          azcore.ClientOptions
}

type Client struct {
	Client                  *azqueue.QueueClient
	AzureQueueStorageConfig AzureQueueStorageConfig
}

func NewClient(config AzureQueueStorageConfig) (*Client, error) {
	if config.ConnectionString == "" {
		return nil, ErrInvalidConnectionString
	}

	client, err := azqueue.NewQueueClientFromConnectionString(config.ConnectionString, config.Queue, &azqueue.ClientOptions{
		ClientOptions: config.Options,
	})

	if err != nil {
		return nil, ErrInvalidConnectionString
	}

	return &Client{
		Client:                  client,
		AzureQueueStorageConfig: config,
	}, nil
}
