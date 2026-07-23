package ports

import (
	"context"
	"io"
	"time"
)

type DBService[T any] interface {
	Insert(ctx context.Context, body T, id *string) (T, error)
	Find(ctx context.Context, id string) (T, error)
	Update(ctx context.Context, id string, fields map[string]any) (T, error)
	Delete(ctx context.Context, id string) error
}

type QueueService interface {
	Enqueue(ctx context.Context, body string) (string, error)
	Dequeue(ctx context.Context) (string, error)
}

type FileService interface {
	Upload(ctx context.Context, fileName string, reader io.Reader) error
	Download(ctx context.Context, fileName string, writer io.Writer) error
	DeleteFile(ctx context.Context, fileName string) error
	GenerateShareUrl(ctx context.Context, fileName string, exp time.Duration) (string, error)
	GenerateUploadUrl(ctx context.Context, fileName string, contentType string, exp time.Duration) (string, error)
}
