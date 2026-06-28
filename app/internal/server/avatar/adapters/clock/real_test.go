package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRealClock_Now(t *testing.T) {
	before := time.Now()
	got := NewRealClock().Now()
	after := time.Now()

	assert.False(t, got.Before(before))
	assert.False(t, got.After(after))
}
