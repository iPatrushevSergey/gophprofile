package clock

import "time"

// RealClock provides current wall clock time.
type RealClock struct{}

// NewRealClock creates a RealClock.
func NewRealClock() *RealClock { return &RealClock{} }

// Now returns the current UTC time.
func (RealClock) Now() time.Time { return time.Now() }
