package converter

import (
	"encoding/json"
	"fmt"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
)

// AvatarUploadedEventToMessage serializes uploaded event payload for RabbitMQ.
func AvatarUploadedEventToMessage(event dto.AvatarUploadedEvent) ([]byte, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("encode avatar uploaded event: %w", err)
	}
	return data, nil
}

// AvatarDeletedEventToMessage serializes deleted event payload for RabbitMQ.
func AvatarDeletedEventToMessage(event dto.AvatarDeletedEvent) ([]byte, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("encode avatar deleted event: %w", err)
	}
	return data, nil
}
