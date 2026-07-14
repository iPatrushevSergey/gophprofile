package usecase

import (
	"context"
	"fmt"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
)

// CollectPeriodicMetrics refreshes polled gauge metrics from current system state.
type CollectPeriodicMetrics struct {
	poolStats    pkgport.PoolStats
	avatarReader appport.AvatarReader
	outboxReader appport.OutboxReader
	metrics      pkgport.Metrics
}

// NewCollectPeriodicMetrics returns the collect periodic metrics use case.
func NewCollectPeriodicMetrics(
	poolStats pkgport.PoolStats,
	avatarReader appport.AvatarReader,
	outboxReader appport.OutboxReader,
	metrics pkgport.Metrics,
) appport.UseCase[struct{}, struct{}] {
	return &CollectPeriodicMetrics{
		poolStats:    poolStats,
		avatarReader: avatarReader,
		outboxReader: outboxReader,
		metrics:      metrics,
	}
}

// Execute refreshes polled gauge metrics from repositories and pool statistics.
func (uc *CollectPeriodicMetrics) Execute(ctx context.Context, _ struct{}) (struct{}, error) {
	stats, err := uc.poolStats.Stats(ctx)
	if err != nil {
		return struct{}{}, fmt.Errorf("read db pool stats: %w", err)
	}
	uc.metrics.ObserveDBPool(stats)

	storageBytes, err := uc.avatarReader.SumCompletedStorageBytes(ctx)
	if err != nil {
		return struct{}{}, fmt.Errorf("sum completed storage bytes: %w", err)
	}
	uc.metrics.SetStorageBytes(storageBytes)

	pending, err := uc.outboxReader.CountPending(ctx)
	if err != nil {
		return struct{}{}, fmt.Errorf("count pending outbox messages: %w", err)
	}
	uc.metrics.SetOutboxPending(pending)

	return struct{}{}, nil
}
