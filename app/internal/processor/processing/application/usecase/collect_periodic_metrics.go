package usecase

import (
	"context"
	"fmt"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
)

// CollectPeriodicMetrics refreshes polled gauge metrics from current system state.
type CollectPeriodicMetrics struct {
	poolStats pkgport.PoolStats
	metrics   pkgport.Metrics
}

// NewCollectPeriodicMetrics returns the collect periodic metrics use case.
func NewCollectPeriodicMetrics(
	poolStats pkgport.PoolStats,
	metrics pkgport.Metrics,
) appport.UseCase[struct{}, struct{}] {
	return &CollectPeriodicMetrics{
		poolStats: poolStats,
		metrics:   metrics,
	}
}

// Execute refreshes polled gauge metrics from pool statistics.
func (uc *CollectPeriodicMetrics) Execute(ctx context.Context, _ struct{}) (struct{}, error) {
	stats, err := uc.poolStats.Stats(ctx)
	if err != nil {
		return struct{}{}, fmt.Errorf("read db pool stats: %w", err)
	}
	uc.metrics.ObserveDBPool(ctx, stats)

	return struct{}{}, nil
}
