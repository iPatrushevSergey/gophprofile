package usecase

import (
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
)

// AvatarUseCasesParams contains dependencies required to build avatar use cases.
type AvatarUseCasesParams struct {
	AvatarRepo              appport.AvatarRepo
	AvatarStorage           appport.AvatarStorage
	EventPublisher          appport.EventPublisher
	OutboxRepo              appport.OutboxRepo
	IDGenerator             appport.IDGenerator
	Transactor              appport.Transactor
	Clock                   appport.Clock
	Logger                  appport.Logger
	OutboxBatchSize         int
	OutboxPublishingTimeout time.Duration
	UploadReservationTTL    time.Duration
}

// AvatarUseCases holds avatar module use cases exposed to the composition root.
type AvatarUseCases struct {
	Upload                     appport.UseCase[dto.UploadAvatarInput, dto.UploadAvatarOutput]
	Get                        appport.UseCase[dto.GetAvatarInput, dto.GetAvatarOutput]
	GetMetadata                appport.UseCase[dto.GetAvatarMetadataInput, dto.AvatarMetadataOutput]
	ListByUser                 appport.UseCase[dto.ListUserAvatarsInput, dto.ListUserAvatarsOutput]
	Delete                     appport.UseCase[dto.DeleteAvatarInput, struct{}]
	Health                     appport.UseCase[struct{}, dto.HealthCheckOutput]
	ExpireUploadingAvatars     appport.UseCase[struct{}, struct{}]
	PublishPendingOutboxEvents appport.UseCase[struct{}, struct{}]
}

// NewAvatarUseCases builds avatar module use cases.
func NewAvatarUseCases(p AvatarUseCasesParams) *AvatarUseCases {
	return &AvatarUseCases{
		Upload: NewUploadAvatar(
			p.AvatarRepo,
			p.AvatarStorage,
			p.OutboxRepo,
			p.Transactor,
			p.IDGenerator,
			p.Clock,
		),
		Get:         NewGetAvatar(p.AvatarRepo, p.AvatarStorage),
		GetMetadata: NewGetAvatarMetadata(p.AvatarRepo),
		ListByUser:  NewListUserAvatars(p.AvatarRepo),
		Delete: NewDeleteAvatar(
			p.AvatarRepo,
			p.AvatarRepo,
			p.OutboxRepo,
			p.Transactor,
			p.IDGenerator,
			p.Clock,
		),
		Health: NewHealthCheck(
			p.AvatarRepo,
			p.AvatarStorage,
			p.EventPublisher,
		),
		ExpireUploadingAvatars: NewExpireUploadingAvatars(
			p.AvatarRepo,
			p.AvatarRepo,
			p.AvatarStorage,
			p.Clock,
			p.Logger,
			p.UploadReservationTTL,
		),
		PublishPendingOutboxEvents: NewPublishPendingOutboxEvents(
			p.OutboxRepo,
			p.OutboxRepo,
			p.EventPublisher,
			p.Clock,
			p.OutboxBatchSize,
			p.OutboxPublishingTimeout,
		),
	}
}
