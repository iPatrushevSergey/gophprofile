package converter

import (
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/rabbitmq/model"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.4 gen .

// goverter:converter
// goverter:output:file generated/message.go
type MessageConverter interface {
	AvatarUploadedEventDtoToAvatarUploadedEventModel(source dto.AvatarUploadedEvent) model.AvatarUploadedEvent
	AvatarDeletedEventDtoToAvatarDeletedEventModel(source dto.AvatarDeletedEvent) model.AvatarDeletedEvent
}
