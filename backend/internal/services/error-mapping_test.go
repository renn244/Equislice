package services

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue/v2/queueerror"
)

func responseError(code string, statusCode int) error {
	return &azcore.ResponseError{ErrorCode: code, StatusCode: statusCode}
}

func TestMapBlobError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"missing container", responseError(string(bloberror.ContainerNotFound), http.StatusNotFound), ErrContainerNotFound},
		{"missing blob", responseError(string(bloberror.BlobNotFound), http.StatusNotFound), ErrBlobNotFound},
		{"invalid input", responseError(string(bloberror.InvalidInput), http.StatusBadRequest), ErrInvalidInput},
		{"access denied", responseError("AuthorizationFailure", http.StatusForbidden), ErrBlobAccessDenied},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !errors.Is(mapBlobError(test.err), test.want) {
				t.Fatalf("expected %v", test.want)
			}
		})
	}
}

func TestMapQueueError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"missing queue", responseError(string(queueerror.QueueNotFound), http.StatusNotFound), ErrQueueNotFound},
		{"missing message", responseError(string(queueerror.MessageNotFound), http.StatusNotFound), ErrMessageNotFound},
		{"invalid input", responseError(string(queueerror.MessageTooLarge), http.StatusBadRequest), ErrQueueInvalidInput},
		{"access denied", responseError("AuthorizationFailure", http.StatusForbidden), ErrQueueAccessDenied},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !errors.Is(mapQueueError(test.err), test.want) {
				t.Fatalf("expected %v", test.want)
			}
		})
	}
}

func TestMapTableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"missing table", responseError(string(aztables.TableNotFound), http.StatusNotFound), ErrTableNotFound},
		{"missing entity", responseError(string(aztables.EntityNotFound), http.StatusNotFound), ErrEntityNotFound},
		{"invalid input", responseError(string(aztables.InvalidInput), http.StatusBadRequest), ErrTableInvalidInput},
		{"access denied", responseError("AuthorizationFailure", http.StatusForbidden), ErrTableAccessDenied},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !errors.Is(mapTableError(test.err), test.want) {
				t.Fatalf("expected %v", test.want)
			}
		})
	}
}
