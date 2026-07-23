package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

var (
	ErrContainerNotFound = errors.New("container not found")
	ErrBlobNotFound      = errors.New("blob not found")
	ErrInvalidInput      = errors.New("invalid input parameters")
	ErrBlobAccessDenied  = errors.New("blob access denied")
)

type FileStorageService struct {
	BlobStorageClient *azblob.Client
	StorageName       string
}

func NewFileStorageService(blobStorageClient *azblob.Client, storageName string) *FileStorageService {
	return &FileStorageService{
		BlobStorageClient: blobStorageClient,
		StorageName:       storageName,
	}
}

func (s *FileStorageService) Upload(ctx context.Context, fileName string, reader io.Reader) error {
	_, err := s.BlobStorageClient.UploadStream(ctx, s.StorageName, fileName, reader, nil)
	return mapBlobError(err)
}

func (s *FileStorageService) Download(ctx context.Context, fileName string, writer io.Writer) error {
	blobResponse, err := s.BlobStorageClient.DownloadStream(ctx, s.StorageName, fileName, nil)
	if err != nil {
		return mapBlobError(err)
	}
	defer blobResponse.Body.Close()

	_, err = io.Copy(writer, blobResponse.Body)
	if err != nil {
		return err
	}

	return nil
}

func (s *FileStorageService) DeleteFile(ctx context.Context, fileName string) error {
	_, err := s.BlobStorageClient.DeleteBlob(ctx, s.StorageName, fileName, nil)
	return mapBlobError(err)
}

func mapBlobError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case bloberror.HasCode(err, bloberror.ContainerNotFound):
		return fmt.Errorf("%w: %v", ErrContainerNotFound, err)
	case bloberror.HasCode(err, bloberror.BlobNotFound):
		return fmt.Errorf("%w: %v", ErrBlobNotFound, err)
	case bloberror.HasCode(err, bloberror.InvalidInput, bloberror.InvalidResourceName, bloberror.OutOfRangeInput, bloberror.RequestBodyTooLarge):
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	case bloberror.HasCode(err, bloberror.AuthenticationFailed, bloberror.AuthorizationFailure, bloberror.AuthorizationPermissionMismatch):
		return fmt.Errorf("%w: %v", ErrBlobAccessDenied, err)
	default:
		return fmt.Errorf("azure blob operation: %w", err)
	}
}

func (s *FileStorageService) GenerateShareUrl(ctx context.Context, fileName string, exp time.Duration) (string, error) {
	blobClient := s.BlobStorageClient.ServiceClient().NewContainerClient(s.StorageName).NewBlobClient(fileName)
	permissions := sas.BlobPermissions{
		Read: true,
	}

	expiry := time.Now().Add(exp)
	sasUrl, err := blobClient.GetSASURL(permissions, expiry, nil)
	if err != nil {
		return "", err
	}

	return sasUrl, nil
}

func (s *FileStorageService) GenerateUploadUrl(ctx context.Context, fileName string, contentType string, exp time.Duration) (string, error) {
	blobClient := s.BlobStorageClient.ServiceClient().NewContainerClient(s.StorageName).NewBlobClient(fileName)

	sasPermissions := sas.BlobPermissions{Create: true, Write: true}
	expiry := time.Now().Add(exp)

	sasUrl, err := blobClient.GetSASURL(sasPermissions, expiry, nil)
	if err != nil {
		return "", err
	}

	return sasUrl, nil
}
