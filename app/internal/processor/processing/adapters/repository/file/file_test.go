package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRepository_MarkAlive_emptyPath(t *testing.T) {
	t.Parallel()

	repo := NewFileRepository()
	err := repo.MarkAlive(context.Background(), "  ", time.Now())
	require.Error(t, err)
}

func TestFileRepository_MarkAlive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "health")
	repo := NewFileRepository()

	updatedAt := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	require.NoError(t, repo.MarkAlive(context.Background(), path, updatedAt))
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.False(t, info.IsDir())
	assert.True(t, info.ModTime().Equal(updatedAt) || info.ModTime().Truncate(time.Second).Equal(updatedAt.Truncate(time.Second)))

	later := updatedAt.Add(time.Minute)
	require.NoError(t, repo.MarkAlive(context.Background(), path, later))
	info, err = os.Stat(path)
	require.NoError(t, err)
	assert.True(t, !info.ModTime().Before(later.Add(-time.Second)))
}
