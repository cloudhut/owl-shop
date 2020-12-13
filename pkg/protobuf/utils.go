package protobuf

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func NewTimestamp(time *time.Time) *timestamppb.Timestamp {
	if time == nil {
		return nil
	}

	return timestamppb.New(*time)
}
