package core_grpc_pbtime

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}

	return ts.AsTime()
}

func ToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}

	t := ts.AsTime()

	return &t
}
