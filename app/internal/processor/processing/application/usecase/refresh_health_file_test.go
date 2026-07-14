package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port/mocks"
)

const testHealthFilePath = "/tmp/health"

func TestRefreshHealthFile_Execute(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	ctrl := gomock.NewController(t)
	files := mocks.NewMockFileRepo(ctrl)
	clock := mocks.NewMockClock(ctrl)
	clock.EXPECT().Now().Return(now)
	files.EXPECT().MarkAlive(gomock.Any(), testHealthFilePath, now).Return(nil)

	uc := NewRefreshHealthFile(files, clock, testHealthFilePath)
	_, err := uc.Execute(context.Background(), struct{}{})
	require.NoError(t, err)
}

func TestRefreshHealthFile_Execute_error(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	ctrl := gomock.NewController(t)
	files := mocks.NewMockFileRepo(ctrl)
	clock := mocks.NewMockClock(ctrl)
	clock.EXPECT().Now().Return(now)
	files.EXPECT().MarkAlive(gomock.Any(), testHealthFilePath, now).Return(errors.New("disk full"))

	uc := NewRefreshHealthFile(files, clock, testHealthFilePath)
	_, err := uc.Execute(context.Background(), struct{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mark health file alive")
}

func TestRefreshHealthFile_Execute_emptyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	files := mocks.NewMockFileRepo(ctrl)
	clock := mocks.NewMockClock(ctrl)

	uc := NewRefreshHealthFile(files, clock, "  ")
	_, err := uc.Execute(context.Background(), struct{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "health file path is required")
}
