package migrate

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationsGophprofileDir(t *testing.T) {
	dir := MigrationsGophprofileDir()
	_, err := os.Stat(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, dir)
}
