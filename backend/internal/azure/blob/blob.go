package blob

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var (
	ErrInvalidConnectionString = errors.New("invalid connection string")
)

type AzureBlobStorageConfig struct {
	Container        string
	ConnectionString string
}

type Client struct {
	Client                 *azblob.Client
	AzureBlobStorageConfig AzureBlobStorageConfig
}

func NewClient(config AzureBlobStorageConfig) (*Client, error) {
	if config.ConnectionString == "" {
		return nil, ErrInvalidConnectionString
	}

	client, err := azblob.NewClientFromConnectionString(config.ConnectionString, nil)
	if err != nil {
		return nil, ErrInvalidConnectionString
	}

	return &Client{
		Client:                 client,
		AzureBlobStorageConfig: config,
	}, nil
}
