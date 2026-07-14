package generator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIDGenerator_NewID(t *testing.T) {
	gen := NewIDGenerator()

	id, err := gen.NewID()
	require.NoError(t, err)
	require.NotEmpty(t, id)

	id2, err := gen.NewID()
	require.NoError(t, err)
	require.NotEqual(t, id, id2)
}
