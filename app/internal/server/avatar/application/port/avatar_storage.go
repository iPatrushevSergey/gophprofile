package port

import "context"

// AvatarStorage stores avatar file payloads in object storage.
type AvatarStorage interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}
