// Package port defines the presentation-layer contract for avatar use cases.
package port

import (
	appdto "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
)

// AvatarUseCases provides avatar use cases to HTTP handlers.
type AvatarUseCases interface {
	UploadUseCase() appport.UseCase[appdto.UploadAvatarInput, appdto.UploadAvatarOutput]
	GetUseCase() appport.UseCase[appdto.GetAvatarInput, appdto.GetAvatarOutput]
	GetMetadataUseCase() appport.UseCase[appdto.GetAvatarMetadataInput, appdto.AvatarMetadataOutput]
	DeleteUseCase() appport.UseCase[appdto.DeleteAvatarInput, struct{}]
	HealthUseCase() appport.UseCase[appdto.HealthCheckInput, appdto.HealthCheckOutput]
}
