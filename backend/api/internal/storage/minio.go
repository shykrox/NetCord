package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOObjectStore struct {
	client *minio.Client
}

func NewMinIOObjectStore(endpoint, accessKey, secretKey string) (*MinIOObjectStore, error) {
	endpoint = strings.TrimSpace(endpoint)
	secure := false

	if strings.HasPrefix(endpoint, "https://") {
		secure = true
		endpoint = strings.TrimPrefix(endpoint, "https://")
	} else {
		endpoint = strings.TrimPrefix(endpoint, "http://")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	return &MinIOObjectStore{client: client}, nil
}

func (s *MinIOObjectStore) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check minio bucket: %w", err)
	}
	if exists {
		return nil
	}

	if err := s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create minio bucket: %w", err)
	}
	return nil
}

func (s *MinIOObjectStore) Put(ctx context.Context, bucket, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("put minio object: %w", err)
	}
	return nil
}

func (s *MinIOObjectStore) Get(ctx context.Context, bucket, objectKey string) (io.ReadCloser, ObjectInfo, error) {
	object, err := s.client.GetObject(ctx, bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, ObjectInfo{}, fmt.Errorf("get minio object: %w", err)
	}

	stat, err := object.Stat()
	if err != nil {
		_ = object.Close()
		return nil, ObjectInfo{}, fmt.Errorf("stat minio object: %w", err)
	}

	return object, ObjectInfo{
		Size:        stat.Size,
		ContentType: stat.ContentType,
	}, nil
}
