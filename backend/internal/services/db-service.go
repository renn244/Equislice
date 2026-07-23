package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/google/uuid"
)

var (
	ErrTableNotFound     = errors.New("table not found")
	ErrEntityNotFound    = errors.New("table entity not found")
	ErrTableInvalidInput = errors.New("invalid table input")
	ErrTableAccessDenied = errors.New("table access denied")
)

type TableStorageService[T any] struct {
	TableStorageClient *aztables.Client
	TableName          string
	PartitionKey       string
}

func NewTableStorageService[T any](tableStorageClient *aztables.Client, tableName string, partitionKey string) *TableStorageService[T] {
	return &TableStorageService[T]{
		TableStorageClient: tableStorageClient,
		TableName:          tableName,
		PartitionKey:       partitionKey,
	}
}

func (s *TableStorageService[T]) marshalEntity(id string, body T) ([]byte, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	var merged map[string]any
	if err := json.Unmarshal(bodyBytes, &merged); err != nil {
		return nil, err
	}

	merged["PartitionKey"] = s.PartitionKey
	merged["RowKey"] = id

	return json.Marshal(merged)
}

func (s *TableStorageService[T]) generateId() (string, error) {
	v4, err := uuid.NewRandom()
	if err != nil {
		return "", errors.New("failed to generate v4 uuid")
	}

	return v4.String(), nil
}

func (s *TableStorageService[T]) Insert(ctx context.Context, body T, id *string) (T, error) {
	if id == nil {
		jobId, err := s.generateId()
		if err != nil {
			return body, err
		}

		id = &jobId
	}

	entityBytes, err := s.marshalEntity(*id, body)
	if err != nil {
		return body, err
	}

	_, err = s.TableStorageClient.AddEntity(ctx, entityBytes, nil)
	if err != nil {
		return body, mapTableError(err)
	}

	return body, nil
}

func (s *TableStorageService[T]) Find(ctx context.Context, id string) (T, error) {
	var result T

	resp, err := s.TableStorageClient.GetEntity(ctx, s.PartitionKey, id, nil)
	if err != nil {
		return result, mapTableError(err)
	}

	err = json.Unmarshal(resp.Value, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (s *TableStorageService[T]) Update(ctx context.Context, id string, fields map[string]any) (T, error) {
	var result T

	existing, err := s.Find(ctx, id)
	if err != nil {
		return result, err
	}

	existingBytes, err := json.Marshal(existing)
	if err != nil {
		return result, err
	}

	var merged map[string]any
	err = json.Unmarshal(existingBytes, &merged)
	if err != nil {
		return result, err
	}

	for k, v := range fields {
		merged[k] = v
	}

	merged["PartitionKey"] = s.PartitionKey
	merged["RowKey"] = id

	updatedBytes, err := json.Marshal(merged)
	if err != nil {
		return result, err
	}

	_, err = s.TableStorageClient.UpdateEntity(ctx, updatedBytes, nil)
	if err != nil {
		return result, mapTableError(err)
	}

	err = json.Unmarshal(updatedBytes, &result)
	return result, err
}

func (s *TableStorageService[T]) Delete(ctx context.Context, id string) error {
	_, err := s.TableStorageClient.DeleteEntity(ctx, s.PartitionKey, id, nil)
	return mapTableError(err)
}

func mapTableError(err error) error {
	if err == nil {
		return nil
	}

	var responseError *azcore.ResponseError
	if errors.As(err, &responseError) {
		switch aztables.TableErrorCode(responseError.ErrorCode) {
		case aztables.TableNotFound, aztables.ResourceNotFound:
			return fmt.Errorf("%w: %v", ErrTableNotFound, err)
		case aztables.EntityNotFound:
			return fmt.Errorf("%w: %v", ErrEntityNotFound, err)
		case aztables.InvalidInput, aztables.OutOfRangeInput, aztables.EntityTooLarge, aztables.InvalidValueType, aztables.PropertyNameInvalid, aztables.PropertyNameTooLong, aztables.PropertyValueTooLarge, aztables.TooManyProperties:
			return fmt.Errorf("%w: %v", ErrTableInvalidInput, err)
		case "AuthenticationFailed", "AuthorizationFailure", "AuthorizationPermissionMismatch":
			return fmt.Errorf("%w: %v", ErrTableAccessDenied, err)
		}

		if responseError.StatusCode == http.StatusUnauthorized || responseError.StatusCode == http.StatusForbidden {
			return fmt.Errorf("%w: %v", ErrTableAccessDenied, err)
		}
	}

	return fmt.Errorf("azure table operation: %w", err)
}
