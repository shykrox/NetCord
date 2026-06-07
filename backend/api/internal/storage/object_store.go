package storage

import (
	"context"
	"io"
)

type ObjectInfo struct {
	Size        int64
	ContentType string
}

type ObjectStore interface {
	EnsureBucket(ctx context.Context, bucket string) error
	Put(ctx context.Context, bucket, objectKey string, reader io.Reader, size int64, contentType string) error
	Get(ctx context.Context, bucket, objectKey string) (io.ReadCloser, ObjectInfo, error)
}
