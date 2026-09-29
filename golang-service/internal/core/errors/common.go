package core_errors

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrConflict        = errors.New("conflict")
	ErrTimeLimit       = errors.New("wait 1 minute")
	ErrForbidden       = errors.New("forbidden")
)
