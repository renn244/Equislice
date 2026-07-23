package table

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
)

var (
	ErrInvalidConnectionString = errors.New("invalid connection string")
	ErrTableNotFound           = errors.New("table not found")
	ErrEntityNotFound          = errors.New("entity not found")
	ErrInvalidInput            = errors.New("invalid input properties")
)

type AzureTableDataConfig struct {
	Table            string
	ConnectionString string
}

type Client struct {
	Client               *aztables.Client
	AzureTableDataConfig AzureTableDataConfig
}

func NewClient(config AzureTableDataConfig) (*Client, error) {
	if config.ConnectionString == "" {
		return nil, ErrInvalidConnectionString
	}

	serviceClient, err := aztables.NewServiceClientFromConnectionString(config.ConnectionString, nil)
	if err != nil {
		return nil, ErrInvalidConnectionString
	}

	client := serviceClient.NewClient(config.Table)

	return &Client{
		Client:               client,
		AzureTableDataConfig: config,
	}, nil
}
