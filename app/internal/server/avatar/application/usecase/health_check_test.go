package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port/mocks"
)

func TestHealthCheck_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("allComponentsUp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := portmocks.NewMockDatabaseHealth(ctrl)
		storage := portmocks.NewMockStorageHealth(ctrl)
		broker := portmocks.NewMockBrokerHealth(ctrl)

		db.EXPECT().Ping(ctx).Return(nil)
		storage.EXPECT().Ping(ctx).Return(nil)
		broker.EXPECT().Ping(ctx).Return(nil)

		uc := NewHealthCheck(db, storage, broker)
		out, err := uc.Execute(ctx, dto.HealthCheckInput{})
		require.NoError(t, err)
		assert.Equal(t, healthStatusOK, out.Status)
		assert.Equal(t, healthStatusOK, out.Database)
		assert.Equal(t, healthStatusOK, out.Storage)
		assert.Equal(t, healthStatusOK, out.Broker)
	})

	t.Run("databaseDown", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := portmocks.NewMockDatabaseHealth(ctrl)
		storage := portmocks.NewMockStorageHealth(ctrl)
		broker := portmocks.NewMockBrokerHealth(ctrl)

		db.EXPECT().Ping(ctx).Return(errors.New("down"))
		storage.EXPECT().Ping(ctx).Return(nil)
		broker.EXPECT().Ping(ctx).Return(nil)

		uc := NewHealthCheck(db, storage, broker)
		out, err := uc.Execute(ctx, dto.HealthCheckInput{})
		require.NoError(t, err)
		assert.Equal(t, healthStatusError, out.Status)
		assert.Equal(t, healthStatusError, out.Database)
	})
}
