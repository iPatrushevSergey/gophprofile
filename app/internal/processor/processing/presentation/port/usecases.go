// Package port defines the presentation-layer contract for processing use cases.
package port

import (
	appdto "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
)

// ProcessingUseCases provides processing use cases to worker handlers.
type ProcessingUseCases interface {
	SubscribeAvatarEventsUseCase() appport.UseCase[struct{}, <-chan appdto.BrokerMessage]
	ConfirmAvatarEventUseCase() appport.UseCase[appdto.ConfirmAvatarEventInput, struct{}]
	ProcessUploadedUseCase() appport.UseCase[appdto.ProcessUploadedAvatarInput, struct{}]
	PurgeDeletedUseCase() appport.UseCase[appdto.PurgeDeletedAvatarInput, struct{}]
	CollectPeriodicMetricsUseCase() appport.UseCase[struct{}, struct{}]
	RefreshHealthFileUseCase() appport.UseCase[struct{}, struct{}]
}
