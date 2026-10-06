package azure

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/getsentry/sentry-go"
)

func CleanupExistingBlobs(ctx context.Context, blob *azblob.Client, existingBlobs []string) {
	for _, blobName := range existingBlobs {
		err := deleteBlob(ctx, blob, "equirectangular-slice", blobName)
		if err != nil {
			sentry.CaptureException(err)
		}
	}

}

func CreateBlobInstance(connectionString string) (*azblob.Client, error) {
	if connectionString == "" {
		return nil, errors.New("Invalid Connection String")
	}

	client, err := azblob.NewClientFromConnectionString(connectionString, &azblob.ClientOptions{
		ClientOptions: tracingOptions(),
	})
	if err != nil {
		return nil, errors.New("Error Connecting to Blob")
	}

	return client, nil
}

func DownloadBlobStream(ctx context.Context, blob *azblob.Client, containerName string, blobName string) (data []byte, operationErr error) {
	ctx, endSpan := startBlobSpan(ctx, "Blob.Download")
	defer func() { endSpan(operationErr) }()

	response, err := blob.DownloadStream(ctx, containerName, blobName, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	imageResponse, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return imageResponse, nil
}

func GetPropertiesBlob(ctx context.Context, blob *azblob.Client, containerName string, blobName string) (properties *blob.GetPropertiesResponse, operationErr error) {
	ctx, endSpan := startBlobSpan(ctx, "Blob.GetProperties")
	defer func() { endSpan(operationErr) }()

	containerClient := blob.ServiceClient().NewContainerClient(containerName)
	tileClient := containerClient.NewBlockBlobClient(blobName)
	data, err := tileClient.GetProperties(ctx, nil)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func UploadBlobStream(ctx context.Context, blob *azblob.Client, containerName string, blobName string, blobData []byte) (result *azblob.UploadStreamResponse, operationErr error) {
	ctx, endSpan := startBlobSpan(ctx, "Blob.Upload")
	defer func() { endSpan(operationErr) }()

	reader := bytes.NewReader(blobData)
	response, err := blob.UploadStream(ctx, containerName, blobName, reader, nil)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func deleteBlob(ctx context.Context, blob *azblob.Client, containerName string, blobName string) (operationErr error) {
	ctx, endSpan := startBlobSpan(ctx, "Blob.Delete")
	defer func() { endSpan(operationErr) }()

	_, err := blob.DeleteBlob(ctx, containerName, blobName, nil)
	if err != nil {
		return err
	}

	return nil
}
