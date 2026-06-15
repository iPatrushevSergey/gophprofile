package application

import "errors"

var (
	// ErrNotFound is returned when avatar is not found.
	ErrNotFound = errors.New("not found")
	// ErrForbidden is returned when user is not allowed to access avatar.
	ErrForbidden = errors.New("forbidden")
	// ErrBadInput is returned when input validation fails.
	ErrBadInput = errors.New("bad input")
	// ErrStorageUnavailable is returned when object storage is unavailable.
	ErrStorageUnavailable = errors.New("object storage unavailable")
	// ErrBrokerUnavailable is returned when message broker is unavailable.
	ErrBrokerUnavailable = errors.New("message broker unavailable")
)
