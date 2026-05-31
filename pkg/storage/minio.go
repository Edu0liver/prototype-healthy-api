package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOConfig configures the S3-compatible object store.
type MinIOConfig struct {
	Endpoint  string // host:port, e.g. minio:9000 (no scheme)
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	Region    string
}

// MinIOStorage stores objects in an S3-compatible bucket (MinIO).
type MinIOStorage struct {
	client *minio.Client
	bucket string
}

// NewMinIO builds the MinIO-backed storage and ensures the bucket exists.
func NewMinIO(cfg MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("storage: bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
			return nil, fmt.Errorf("storage: make bucket: %w", err)
		}
	}

	return &MinIOStorage{client: client, bucket: cfg.Bucket}, nil
}

// Put stores data under the per-tenant key and returns the object key.
func (s *MinIOStorage) Put(ctx context.Context, companyID uuid.UUID, key string, data []byte) (string, error) {
	objectKey := TenantPrefix(companyID, key)
	if _, err := s.client.PutObject(
		ctx, s.bucket, objectKey, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{},
	); err != nil {
		return "", fmt.Errorf("storage: put: %w", err)
	}
	return objectKey, nil
}

// Get reads the object at the given key.
func (s *MinIOStorage) Get(ctx context.Context, path string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("storage: get: %w", err)
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

// Delete removes the object at the given key. Missing objects are not an error.
func (s *MinIOStorage) Delete(ctx context.Context, path string) error {
	return s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{})
}
