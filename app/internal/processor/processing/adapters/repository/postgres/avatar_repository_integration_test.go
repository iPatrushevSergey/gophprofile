//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgreskit "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/repository/postgres"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/retry"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/testutil"
	avatarpostgres "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/repository/postgres"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	processorvo "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
	serverpostgres "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres"
	serverentity "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

func setupRepos(tb testing.TB) (*serverpostgres.AvatarRepository, *avatarpostgres.AvatarRepository, *pgxpool.Pool) {
	tb.Helper()
	pool := testutil.SetupPostgres(tb)
	tx := postgreskit.NewTransactor(pool, retry.WithMaxRetries(0))
	return serverpostgres.NewAvatarRepository(tx), avatarpostgres.NewAvatarRepository(tx), pool
}

func truncateAvatars(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `TRUNCATE avatars RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func TestAvatarRepository_UpdateProcessingStatusAndComplete(t *testing.T) {
	serverRepo, processorRepo, pool := setupRepos(t)
	truncateAvatars(t, pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	id := uuid.NewString()
	avatar := serverentity.NewAvatar(
		id,
		"user-1",
		"avatar.png",
		"image/png",
		100,
		serverentity.OriginalObjectKey("user-1", id),
		vo.UploadStatusCompleted,
		now,
	)
	require.NoError(t, serverRepo.Create(ctx, avatar))

	require.NoError(t, processorRepo.UpdateProcessingStatus(ctx, id, processorvo.ProcessingStatusProcessing, now))

	thumbnailKeys := map[processorvo.ThumbnailSize]map[processorvo.OutputFormat]string{
		processorvo.ThumbnailSize100: {processorvo.OutputFormatJPEG: "user-1/" + id + "/100x100/jpeg"},
	}
	require.NoError(t, processorRepo.CompleteProcessing(ctx, dto.CompleteProcessingInput{
		AvatarID:        id,
		ThumbnailS3Keys: thumbnailKeys,
		Width:           64,
		Height:          64,
		UpdatedAt:       now,
	}))

	found, err := processorRepo.FindByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, processorvo.ProcessingStatusCompleted, found.ProcessingStatus)
	assert.Equal(t, "user-1/"+id+"/100x100/jpeg", found.ThumbnailS3Keys[processorvo.ThumbnailSize100][processorvo.OutputFormatJPEG])
}
