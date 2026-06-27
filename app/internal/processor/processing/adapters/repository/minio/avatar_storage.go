package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
)

var _ appport.AvatarStorage = (*AvatarStorage)(nil)

// AvatarStorage stores avatar file payloads in MinIO.
type AvatarStorage struct {
	client *minio.Client
	bucket string
}

// NewAvatarStorage creates a MinIO-backed avatar storage adapter.
func NewAvatarStorage(cfg Config) (*AvatarStorage, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("minio endpoint is required")
	}

	if cfg.Bucket == "" {
		return nil, fmt.Errorf("minio bucket is required")
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket exists: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio make bucket: %w", err)
		}
	}

	return &AvatarStorage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Put stores payload bytes under key in the configured bucket.
func (s *AvatarStorage) Put(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := s.client.PutObject(
		ctx,
		s.bucket,
		key,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return fmt.Errorf("minio put object: %w", err)
	}
	return nil
}

// Get loads payload bytes stored under key.
func (s *AvatarStorage) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio get object: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("minio read object: %w", err)
	}
	return data, nil
}

// Delete removes payload stored under key.
func (s *AvatarStorage) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("minio delete object: %w", err)
	}
	return nil
}

// Ping checks object storage connectivity.
func (s *AvatarStorage) Ping(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("minio bucket exists: %w", err)
	}
	return nil
}
