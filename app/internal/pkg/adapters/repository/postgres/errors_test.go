package postgres

import (
	"io"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsRetriable_network(t *testing.T) {
	assert.True(t, IsRetriable(io.EOF))
	assert.False(t, IsRetriable(nil))
}

func TestIsRetriable_pgCodes(t *testing.T) {
	assert.True(t, IsRetriable(&pgconn.PgError{Code: pgerrcode.DeadlockDetected}))
	assert.False(t, IsRetriable(&pgconn.PgError{Code: pgerrcode.UniqueViolation}))
}
