package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue/v2/queueerror"
)

var (
	ErrQueueNotFound     = errors.New("queue not found")
	ErrMessageNotFound   = errors.New("queue message not found")
	ErrQueueEmpty        = errors.New("queue is empty")
	ErrQueueInvalidInput = errors.New("invalid queue input")
	ErrQueueAccessDenied = errors.New("queue access denied")
)

type QueueService struct {
	QueueStorageClient *azqueue.QueueClient
	QueueName          string
}

func NewQueueService(queueStorageClient *azqueue.QueueClient, queueName string) *QueueService {
	return &QueueService{
		QueueStorageClient: queueStorageClient,
		QueueName:          queueName,
	}
}

func (s *QueueService) Enqueue(ctx context.Context, body string) (string, error) {
	_, err := s.QueueStorageClient.EnqueueMessage(ctx, body, nil)
	if err != nil {
		return "", mapQueueError(err)
	}

	return body, nil
}

func (s *QueueService) Dequeue(ctx context.Context) (string, error) {
	queueResponse, err := s.QueueStorageClient.DequeueMessage(ctx, nil)
	if err != nil {
		return "", mapQueueError(err)
	}

	if len(queueResponse.Messages) == 0 {
		return "", ErrQueueEmpty
	}

	return *queueResponse.Messages[0].MessageText, nil
}

func mapQueueError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case queueerror.HasCode(err, queueerror.QueueNotFound, queueerror.ResourceNotFound):
		return fmt.Errorf("%w: %v", ErrQueueNotFound, err)
	case queueerror.HasCode(err, queueerror.MessageNotFound):
		return fmt.Errorf("%w: %v", ErrMessageNotFound, err)
	case queueerror.HasCode(err, queueerror.InvalidInput, queueerror.InvalidResourceName, queueerror.OutOfRangeInput, queueerror.MessageTooLarge, queueerror.RequestBodyTooLarge):
		return fmt.Errorf("%w: %v", ErrQueueInvalidInput, err)
	case queueerror.HasCode(err, queueerror.AuthenticationFailed, queueerror.AuthorizationFailure, queueerror.AuthorizationPermissionMismatch):
		return fmt.Errorf("%w: %v", ErrQueueAccessDenied, err)
	default:
		return fmt.Errorf("azure queue operation: %w", err)
	}
}
