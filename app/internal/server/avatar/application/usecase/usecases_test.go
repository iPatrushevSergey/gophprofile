package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
	pkginmemory "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/repository/inmemory"
	postgresadapter "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/repository/postgres"
	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port/mocks"
)

func TestNewAvatarUseCases_wiresUseCases(t *testing.T) {
	ctrl := gomock.NewController(t)

	uc := NewAvatarUseCases(AvatarUseCasesParams{
		AvatarRepo:              portmocks.NewMockAvatarRepo(ctrl),
		AvatarStorage:           portmocks.NewMockAvatarStorage(ctrl),
		EventPublisher:          portmocks.NewMockEventPublisher(ctrl),
		OutboxRepo:              portmocks.NewMockOutboxRepo(ctrl),
		IDGenerator:             portmocks.NewMockIDGenerator(ctrl),
		Transactor:              pkginmemory.NewTransactor(),
		Clock:                   portmocks.NewMockClock(ctrl),
		Logger:                  logger.NewNopLogger(),
		PoolStats:               postgresadapter.NewNopPoolStats(),
		OutboxBatchSize:         100,
		OutboxPublishingTimeout: 5 * time.Minute,
		UploadReservationTTL:    time.Minute,
	})

	assert.NotNil(t, uc.Upload)
	assert.NotNil(t, uc.Get)
	assert.NotNil(t, uc.GetMetadata)
	assert.NotNil(t, uc.ListByUser)
	assert.NotNil(t, uc.Delete)
	assert.NotNil(t, uc.Health)
	assert.NotNil(t, uc.ExpireUploadingAvatars)
	assert.NotNil(t, uc.PublishPendingOutboxEvents)
	assert.NotNil(t, uc.CollectPeriodicMetrics)
}
