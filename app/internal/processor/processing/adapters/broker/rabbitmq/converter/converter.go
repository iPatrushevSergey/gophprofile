package converter

import (
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/broker/rabbitmq/model"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.4 gen .

// goverter:converter
// goverter:output:file generated/message.go
type MessageConverter interface {
	AvatarUploadedEventModelToAvatarUploadedEventDto(source model.AvatarUploadedEvent) dto.AvatarUploadedEvent
	AvatarDeletedEventModelToAvatarDeletedEventDto(source model.AvatarDeletedEvent) dto.AvatarDeletedEvent
}
