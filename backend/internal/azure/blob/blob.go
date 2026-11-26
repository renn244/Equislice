package blob

import (
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var (
	ErrInvalidConnectionString = errors.New("invalid connection string")
)

type AzureBlobStorageConfig struct {
	Container        string
	ConnectionString string
	Options          azcore.ClientOptions
}

type Client struct {
	Client                 *azblob.Client
	AzureBlobStorageConfig AzureBlobStorageConfig
}

func NewClient(config AzureBlobStorageConfig) (*Client, error) {
	if config.ConnectionString == "" {
		return nil, ErrInvalidConnectionString
	}

	client, err := azblob.NewClientFromConnectionString(config.ConnectionString, &azblob.ClientOptions{
		ClientOptions: config.Options,
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConnectionString, err)
	}

	return &Client{
		Client:                 client,
		AzureBlobStorageConfig: config,
	}, nil
}
