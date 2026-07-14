package application

import "errors"

var (
	// ErrNotFound is returned when avatar is not found.
	ErrNotFound = errors.New("not found")
	// ErrBadInput is returned when input validation fails.
	ErrBadInput = errors.New("bad input")
	// ErrAlreadyProcessed is returned when event is a duplicate.
	ErrAlreadyProcessed = errors.New("already processed")
	// ErrStorageUnavailable is returned when object storage is unavailable.
	ErrStorageUnavailable = errors.New("object storage unavailable")
)
