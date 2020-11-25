package kafka

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kmsg"
)

type TopicLogDirSize struct {
	// TotalSize is the sum of primary log dir size and replica log dir size.
	TotalSize int64

	// LeaderLogDirSize describes the total size of all leaders' replica log dirs.
	LeaderLogDirSize int64

	// ReplicaLogDirSize describes the total size of all replica log dirs.
	ReplicaLogDirSize int64
}

// GetTopicSize returns the topic's log dir size in bytes.
func (s *Service) GetTopicSize(ctx context.Context, topicName string, partitionIDs []int32) (TopicLogDirSize, error) {
	req := kmsg.DescribeLogDirsRequest{
		Topics: []kmsg.DescribeLogDirsRequestTopic{
			{
				Topic:      topicName,
				Partitions: partitionIDs,
			},
		},
	}
	res, err := req.RequestWith(ctx, s.KafkaClient)
	if err != nil {
		return TopicLogDirSize{}, err
	}

	// We have to deduplciate log dir sizes for each partition due to replication. We want to differ between leader
	// and replica log dir size.
	leaderSizeByPartition := make(map[int32]int64)
	replicaSizeByPartition := make(map[int32]int64)
	for _, dir := range res.Dirs {
		err := kerr.ErrorForCode(dir.ErrorCode)
		if err != nil {
			return TopicLogDirSize{}, fmt.Errorf("failed to describe log dir: '%v': %w", dir.Dir, err)
		}
		for _, topic := range dir.Topics {
			for _, partition := range topic.Partitions {
				pID := partition.Partition
				value, exists := leaderSizeByPartition[pID]
				if !exists {
					leaderSizeByPartition[pID] = partition.Size
				}

				if partition.Size > value {
					// THis partition's log dir size is bigger than the currently stored log dir size, so this is
					// more likely the leader's log dir. Move the currently stored value to the replica log dir size.
					replicaSizeByPartition[pID] = value
					leaderSizeByPartition[pID] += partition.Size
				} else {
					leaderSizeByPartition[pID] += partition.Size
				}
			}
		}
	}

	var leaderLogDirSize, replicaLogDirSize int64
	for _, pID := range partitionIDs {
		leaderLogDirSize += leaderSizeByPartition[pID]
		replicaLogDirSize += replicaSizeByPartition[pID]
	}

	return TopicLogDirSize{
		TotalSize:         leaderLogDirSize + replicaLogDirSize,
		LeaderLogDirSize:  leaderLogDirSize,
		ReplicaLogDirSize: replicaLogDirSize,
	}, nil
}
