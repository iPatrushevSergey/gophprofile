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
	avatarRepo     appport.AvatarRepo
	avatarStorage  appport.AvatarStorage
	eventPublisher appport.EventPublisher
}

// NewHealthCheck returns the health check use case.
func NewHealthCheck(
	avatarRepo appport.AvatarRepo,
	avatarStorage appport.AvatarStorage,
	eventPublisher appport.EventPublisher,
) appport.UseCase[struct{}, dto.HealthCheckOutput] {
	return &HealthCheck{
		avatarRepo:     avatarRepo,
		avatarStorage:  avatarStorage,
		eventPublisher: eventPublisher,
	}
}

// Execute checks database, object storage and broker.
func (uc *HealthCheck) Execute(ctx context.Context, _ struct{}) (dto.HealthCheckOutput, error) {
	out := dto.HealthCheckOutput{
		Status:   healthStatusOK,
		Database: healthStatusOK,
		Storage:  healthStatusOK,
		Broker:   healthStatusOK,
	}

	if err := uc.avatarRepo.Ping(ctx); err != nil {
		out.Database = healthStatusError
	}
	if err := uc.avatarStorage.Ping(ctx); err != nil {
		out.Storage = healthStatusError
	}
	if err := uc.eventPublisher.Ping(ctx); err != nil {
		out.Broker = healthStatusError
	}

	if out.Database != healthStatusOK || out.Storage != healthStatusOK || out.Broker != healthStatusOK {
		out.Status = healthStatusError
	}

	return out, nil
}
