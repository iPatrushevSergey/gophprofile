package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoWithRetry_success(t *testing.T) {
	err := DoWithRetry(context.Background(), func() error { return nil })
	assert.NoError(t, err)
}

func TestDoWithRetry_nonRetriable(t *testing.T) {
	want := errors.New("fail")
	err := DoWithRetry(context.Background(), func() error { return want })
	assert.ErrorIs(t, err, want)
}

func TestDoWithRetry_retriesThenSucceeds(t *testing.T) {
	var calls int
	err := DoWithRetry(context.Background(), func() error {
		calls++
		if calls < 2 {
			return errors.New("temporary")
		}
		return nil
	},
		WithMaxRetries(2),
		WithConstantBackoff(time.Millisecond),
		WithRetriableCheck(func(error) bool { return true }),
	)
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}

func TestDoWithRetry_exhausted(t *testing.T) {
	want := errors.New("always")
	err := DoWithRetry(context.Background(), func() error { return want },
		WithMaxRetries(1),
		WithConstantBackoff(time.Millisecond),
		WithRetriableCheck(func(error) bool { return true }),
	)
	assert.ErrorIs(t, err, want)
}
