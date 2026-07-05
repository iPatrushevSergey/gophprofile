package usecase

import (
	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
)

// ProcessingUseCasesParams contains dependencies required to build processing use cases.
type ProcessingUseCasesParams struct {
	AvatarRepo     appport.AvatarRepo
	AvatarStorage  appport.AvatarStorage
	ImageProcessor appport.ImageProcessor
	EventConsumer  appport.EventConsumer
	Clock          appport.Clock
	Tracer         pkgport.Tracer
	Metrics        pkgport.Metrics
}

// ProcessingUseCases holds processing module use cases exposed to the composition root.
type ProcessingUseCases struct {
	SubscribeAvatarEvents appport.UseCase[struct{}, <-chan dto.BrokerMessage]
	ConfirmAvatarEvent    appport.UseCase[dto.ConfirmAvatarEventInput, struct{}]
	ProcessUploaded       appport.UseCase[dto.ProcessUploadedAvatarInput, struct{}]
	PurgeDeleted          appport.UseCase[dto.PurgeDeletedAvatarInput, struct{}]
}

// NewProcessingUseCases builds processing module use cases.
func NewProcessingUseCases(p ProcessingUseCasesParams) *ProcessingUseCases {
	return &ProcessingUseCases{
		SubscribeAvatarEvents: NewSubscribeAvatarEvents(p.EventConsumer),
		ConfirmAvatarEvent:    NewConfirmAvatarEvent(),
		ProcessUploaded: NewProcessUploadedAvatar(
			p.AvatarRepo,
			p.AvatarStorage,
			p.ImageProcessor,
			p.Clock,
			p.Tracer,
			p.Metrics,
		),
		PurgeDeleted: NewPurgeDeletedAvatar(p.AvatarStorage, p.Tracer),
	}
}
