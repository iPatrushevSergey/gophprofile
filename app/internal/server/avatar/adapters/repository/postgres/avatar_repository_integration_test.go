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
	avatarpostgres "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

func setupAvatarRepo(tb testing.TB) (*avatarpostgres.AvatarRepository, *pgxpool.Pool) {
	tb.Helper()
	pool := testutil.SetupPostgres(tb)
	tx := postgreskit.NewTransactor(pool, retry.WithMaxRetries(0))
	return avatarpostgres.NewAvatarRepository(tx), pool
}

func truncateAvatars(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `TRUNCATE avatars RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func completedAvatar(now time.Time) *entity.Avatar {
	id := uuid.NewString()
	return entity.NewAvatar(
		id,
		"user-1",
		"avatar.png",
		"image/png",
		100,
		entity.OriginalObjectKey("user-1", id),
		vo.UploadStatusCompleted,
		now,
	)
}

func TestAvatarRepository_CreateAndFindByID(t *testing.T) {
	repo, pool := setupAvatarRepo(t)
	truncateAvatars(t, pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	avatar := completedAvatar(now)
	require.NoError(t, repo.Create(ctx, avatar))

	found, err := repo.FindByID(ctx, avatar.ID)
	require.NoError(t, err)
	assert.Equal(t, avatar.ID, found.ID)
	assert.Equal(t, vo.UploadStatusCompleted, found.UploadStatus)
}

func TestAvatarRepository_MarkUploadCompleted(t *testing.T) {
	repo, pool := setupAvatarRepo(t)
	truncateAvatars(t, pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	id := uuid.NewString()
	avatar := entity.NewAvatar(
		id,
		"user-1",
		"avatar.png",
		"image/png",
		100,
		entity.OriginalObjectKey("user-1", id),
		vo.UploadStatusUploading,
		now,
	)
	require.NoError(t, repo.Create(ctx, avatar))
	require.NoError(t, repo.MarkUploadCompleted(ctx, id, now))

	found, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, vo.UploadStatusCompleted, found.UploadStatus)
}

func TestAvatarRepository_ListByUserID(t *testing.T) {
	repo, pool := setupAvatarRepo(t)
	truncateAvatars(t, pool)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	require.NoError(t, repo.Create(ctx, completedAvatar(now)))
	require.NoError(t, repo.Create(ctx, completedAvatar(now)))

	items, err := repo.ListByUserID(ctx, "user-1")
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestAvatarRepository_FindByID_NotFound(t *testing.T) {
	repo, pool := setupAvatarRepo(t)
	truncateAvatars(t, pool)

	_, err := repo.FindByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	assert.ErrorIs(t, err, application.ErrNotFound)
}
