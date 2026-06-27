package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port/mocks"
)

func TestNewProcessingUseCases_wiresUseCases(t *testing.T) {
	ctrl := gomock.NewController(t)

	uc := NewProcessingUseCases(ProcessingUseCasesParams{
		AvatarRepo:    portmocks.NewMockAvatarRepo(ctrl),
		AvatarStorage: portmocks.NewMockAvatarStorage(ctrl),
		ImageResizer:  portmocks.NewMockImageResizer(ctrl),
		EventConsumer: portmocks.NewMockEventConsumer(ctrl),
		Clock:         portmocks.NewMockClock(ctrl),
	})

	assert.NotNil(t, uc.SubscribeAvatarEvents)
	assert.NotNil(t, uc.ConfirmAvatarEvent)
	assert.NotNil(t, uc.ProcessUploaded)
	assert.NotNil(t, uc.PurgeDeleted)
}
