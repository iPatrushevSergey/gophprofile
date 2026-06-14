package apputil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApply(t *testing.T) {
	type cfg struct{ n int }
	var c cfg
	Apply(&c, func(c *cfg) { c.n = 1 }, nil, func(c *cfg) { c.n = 2 })
	assert.Equal(t, 2, c.n)
}
