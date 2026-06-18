package port

import "context"

// DatabaseHealth checks PostgreSQL availability.
type DatabaseHealth interface {
	Ping(ctx context.Context) error
}

// StorageHealth checks object storage availability.
type StorageHealth interface {
	Ping(ctx context.Context) error
}

// BrokerHealth checks message broker availability.
type BrokerHealth interface {
	Ping(ctx context.Context) error
}
