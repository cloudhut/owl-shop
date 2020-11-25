package kafka

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"time"
)

// TopicMessage count tries to return the number of kafka messages in that topic. Depending on the configuration
// this might be as simple as returning the delta between low and high watermark (delete cleanup policy), but is more
// complex for compacted topics.
// For compacted topics the number of messages will be estimated by dividing the log dir size with the average
// message size.
func (s *Service) GetTopicMessageCount(ctx context.Context, topicName string) (int64, error) {
	// 1. Figure out cleanup policy
	req := kmsg.DescribeConfigsRequest{
		Resources: []kmsg.DescribeConfigsRequestResource{
			{
				ResourceType: 2, // TOPICS
				ResourceName: topicName,
				ConfigNames:  []string{"cleanup.policy"},
			},
		},
	}
	res, err := req.RequestWith(ctx, s.KafkaClient)
	if err != nil {
		return 0, fmt.Errorf("failed to request topic config: %w", err)
	}

	if len(res.Resources) != 1 {
		return 0, fmt.Errorf("expected 1 resource, but actually got '%v' resources", len(res.Resources))
	}

	topicResource := res.Resources[0]
	if len(topicResource.Configs) != 1 {
		return 0, fmt.Errorf("expected 1 config resource, but actually got '%v' resources", len(topicResource.Configs))
	}

	cleanupPolicy := topicResource.Configs[0].Value
	if cleanupPolicy == nil {
		return 0, fmt.Errorf("returned config value is null")
	}

	// 2. Return watermark delta if cleanup policy is set to delete or the number of total messages can't be more than 5
	topicMetadata, err := s.GetTopicMetadata(ctx, topicName)
	if err != nil {
		return 0, fmt.Errorf("failed to get topic metadata to fetch partition marks: %w", err)
	}
	err = kerr.ErrorForCode(topicMetadata.ErrorCode)
	if err != nil {
		return 0, fmt.Errorf("failed to get topic metadata to fetch partition marks: %w", err)
	}

	partitionIDs := make([]int32, len(topicMetadata.Partitions))
	for _, partition := range topicMetadata.Partitions {
		partitionIDs = append(partitionIDs, partition.Partition)
	}

	totalMessageCount := int64(0)
	marks, err := s.GetPartitionMarks(ctx, topicName, partitionIDs)
	for _, mark := range marks {
		totalMessageCount += mark.High - mark.Low
	}

	if *cleanupPolicy == "delete" || totalMessageCount < 5 {
		return totalMessageCount, nil
	}

	// 3. Return estimated number of messages based on partition log dir size and average message size
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	avgMessageSize, err := s.GetAverageMessageSize(ctxWithTimeout, topicName)
	if err != nil {
		return 0, fmt.Errorf("failed to get average message size: %w", err)
	}

	topicSize, err := s.GetTopicSize(ctx, topicName, partitionIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to get topic log dirs: %w", err)
	}

	estimatedMessageCount := topicSize.LeaderLogDirSize / avgMessageSize
	return estimatedMessageCount, nil
}

// GetAverageMessageSize returns the average message size in a given topic in bytes.
func (s *Service) GetAverageMessageSize(ctx context.Context, topicName string) (int64, error) {
	// 1. Consume a few messages
	client, err := s.NewKafkaClient()
	if err != nil {
		return 0, fmt.Errorf("failed to create new kafka client: %w", err)
	}

	offset := kgo.NewOffset().At(0)
	consumeOpts := kgo.ConsumeTopics(offset, topicName)
	client.AssignPartitions(consumeOpts)

	// One poll should be enough, might be necessary to tune?
	fetches := client.PollFetches(ctx)
	errors := fetches.Errors()
	if len(errors) > 0 {
		return 0, fmt.Errorf("failed to poll messages: %w", errors[0].Err)
	}
	iter := fetches.RecordIter()
	var messageCount, totalMessageSize int64
	for !iter.Done() {
		record := iter.Next()
		messageCount++
		totalMessageSize += int64(len(record.Value))
	}
	avgSize := totalMessageSize / messageCount

	return avgSize, nil
}
