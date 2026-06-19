package usecase

import (
	"context"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
)

const (
	healthStatusOK    = "ok"
	healthStatusError = "error"
)

// HealthCheck reports availability of infrastructure dependencies.
type HealthCheck struct {
	databaseHealth appport.DatabaseHealth
	storageHealth  appport.StorageHealth
	brokerHealth   appport.BrokerHealth
}

// NewHealthCheck returns the health check use case.
func NewHealthCheck(
	databaseHealth appport.DatabaseHealth,
	storageHealth appport.StorageHealth,
	brokerHealth appport.BrokerHealth,
) appport.UseCase[dto.HealthCheckInput, dto.HealthCheckOutput] {
	return &HealthCheck{
		databaseHealth: databaseHealth,
		storageHealth:  storageHealth,
		brokerHealth:   brokerHealth,
	}
}

// Execute checks database, object storage and broker.
func (uc *HealthCheck) Execute(ctx context.Context, _ dto.HealthCheckInput) (dto.HealthCheckOutput, error) {
	out := dto.HealthCheckOutput{
		Status:   healthStatusOK,
		Database: healthStatusOK,
		Storage:  healthStatusOK,
		Broker:   healthStatusOK,
	}

	if err := uc.databaseHealth.Ping(ctx); err != nil {
		out.Database = healthStatusError
	}
	if err := uc.storageHealth.Ping(ctx); err != nil {
		out.Storage = healthStatusError
	}
	if err := uc.brokerHealth.Ping(ctx); err != nil {
		out.Broker = healthStatusError
	}

	if out.Database != healthStatusOK || out.Storage != healthStatusOK || out.Broker != healthStatusOK {
		out.Status = healthStatusError
	}

	return out, nil
}
