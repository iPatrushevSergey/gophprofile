package handler

import (
	"context"

	appdto "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
	presport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/port"
)

type stubUseCase[In, Out any] struct {
	out Out
	err error
}

func (s stubUseCase[In, Out]) Execute(_ context.Context, _ In) (Out, error) {
	return s.out, s.err
}

type stubAvatarUseCases struct {
	upload          appport.UseCase[appdto.UploadAvatarInput, appdto.UploadAvatarOutput]
	get             appport.UseCase[appdto.GetAvatarInput, appdto.GetAvatarOutput]
	getMetadata     appport.UseCase[appdto.GetAvatarMetadataInput, appdto.AvatarMetadataOutput]
	listByUser      appport.UseCase[appdto.ListUserAvatarsInput, appdto.ListUserAvatarsOutput]
	delete          appport.UseCase[appdto.DeleteAvatarInput, struct{}]
	health          appport.UseCase[struct{}, appdto.HealthCheckOutput]
	expireUploading appport.UseCase[struct{}, struct{}]
	publishOutbox   appport.UseCase[struct{}, struct{}]
}

func (s stubAvatarUseCases) UploadUseCase() appport.UseCase[appdto.UploadAvatarInput, appdto.UploadAvatarOutput] {
	return s.upload
}

func (s stubAvatarUseCases) GetUseCase() appport.UseCase[appdto.GetAvatarInput, appdto.GetAvatarOutput] {
	return s.get
}

func (s stubAvatarUseCases) GetMetadataUseCase() appport.UseCase[appdto.GetAvatarMetadataInput, appdto.AvatarMetadataOutput] {
	return s.getMetadata
}

func (s stubAvatarUseCases) ListByUserUseCase() appport.UseCase[appdto.ListUserAvatarsInput, appdto.ListUserAvatarsOutput] {
	return s.listByUser
}

func (s stubAvatarUseCases) DeleteUseCase() appport.UseCase[appdto.DeleteAvatarInput, struct{}] {
	return s.delete
}

func (s stubAvatarUseCases) HealthUseCase() appport.UseCase[struct{}, appdto.HealthCheckOutput] {
	return s.health
}

func (s stubAvatarUseCases) ExpireUploadingAvatarsUseCase() appport.UseCase[struct{}, struct{}] {
	return s.expireUploading
}

func (s stubAvatarUseCases) PublishPendingOutboxEventsUseCase() appport.UseCase[struct{}, struct{}] {
	return s.publishOutbox
}

var _ presport.AvatarUseCases = stubAvatarUseCases{}
