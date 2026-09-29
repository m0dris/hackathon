package core_errors

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FromGRPC(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("call db microservice: %w", err)
	}

	var sentinel error

	switch st.Code() {
	case codes.NotFound:
		sentinel = ErrNotFound
	case codes.AlreadyExists:
		sentinel = ErrConflict
	case codes.FailedPrecondition:
		sentinel = ErrConflict
	case codes.InvalidArgument:
		sentinel = ErrInvalidArgument
	case codes.PermissionDenied:
		sentinel = ErrForbidden
	case codes.ResourceExhausted:
		sentinel = ErrTimeLimit
	default:
		return fmt.Errorf("call db microservice: %w", err)
	}

	return errors.Join(sentinel, fmt.Errorf("call db microservice: %w", err))
}
