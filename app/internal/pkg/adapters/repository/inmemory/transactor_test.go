package inmemory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactor_RunInTransaction(t *testing.T) {
	tx := NewTransactor()
	called := false
	err := tx.RunInTransaction(context.Background(), func(ctx context.Context) error {
		called = true
		assert.NotNil(t, ctx)
		return nil
	})
	require.NoError(t, err)
	assert.True(t, called)
}
