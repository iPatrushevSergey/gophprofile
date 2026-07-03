package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	pkgportmocks "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port/mocks"
	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port/mocks"
)

func TestNewProcessingUseCases_wiresUseCases(t *testing.T) {
	ctrl := gomock.NewController(t)

	uc := NewProcessingUseCases(ProcessingUseCasesParams{
		AvatarRepo:     portmocks.NewMockAvatarRepo(ctrl),
		AvatarStorage:  portmocks.NewMockAvatarStorage(ctrl),
		ImageProcessor: portmocks.NewMockImageProcessor(ctrl),
		EventConsumer:  portmocks.NewMockEventConsumer(ctrl),
		Clock:          portmocks.NewMockClock(ctrl),
		Tracer:         pkgportmocks.NewMockTracer(ctrl),
	})

	assert.NotNil(t, uc.SubscribeAvatarEvents)
	assert.NotNil(t, uc.ConfirmAvatarEvent)
	assert.NotNil(t, uc.ProcessUploaded)
	assert.NotNil(t, uc.PurgeDeleted)
}
