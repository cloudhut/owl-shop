package fake

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func newProtoTimestampFromTimePtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
