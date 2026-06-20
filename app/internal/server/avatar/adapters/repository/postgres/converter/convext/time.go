package convext

import "time"

// CopyTime maps time.Time for goverter (required: time.Time has unexported fields).
func CopyTime(v time.Time) time.Time {
	return v
}

// CopyTimePtr maps *time.Time for goverter (required: time.Time has unexported fields).
func CopyTimePtr(v *time.Time) *time.Time {
	return v
}
