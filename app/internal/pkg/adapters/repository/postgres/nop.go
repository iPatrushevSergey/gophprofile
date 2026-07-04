package postgres

import (
	"context"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
)

// NopPoolStats provides empty pool statistics for tests.
type NopPoolStats struct{}

// NewNopPoolStats returns a no-op pool stats implementation for tests.
func NewNopPoolStats() *NopPoolStats {
	return &NopPoolStats{}
}

var _ pkgport.PoolStats = (*NopPoolStats)(nil)

// Stats returns empty pool statistics.
func (NopPoolStats) Stats(context.Context) (pkgport.DBPoolStats, error) {
	return pkgport.DBPoolStats{}, nil
}
