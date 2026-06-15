package clock

import "time"

// RealClock provides current wall clock time.
type RealClock struct{}

func NewRealClock() *RealClock { return &RealClock{} }

func (RealClock) Now() time.Time { return time.Now() }

