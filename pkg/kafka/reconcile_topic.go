package kafka

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

// ReconcileTopic ensures a topic with the given parameters exist in the target cluster.
func ReconcileTopic(
	ctx context.Context,
	kafkaClient *kgo.Client,
	topicName string,
	partitions int32,
	replicationFactor int16,
	configs map[string]*string,
) error {
	adminClient := kadm.NewClient(kafkaClient)
	topicDetails, err := adminClient.ListTopics(ctx, topicName)
	if err != nil {
		return fmt.Errorf("failed to get metadata to check if topic exists: %w", err)
	}

	if !topicDetails.Has(topicName) {
		createTopicsRes, err := adminClient.CreateTopics(
			ctx,
			partitions,
			replicationFactor,
			configs,
			topicName)
		if err != nil {
			return fmt.Errorf("failed to create topic: %w", err)
		}

		topicRes, exists := createTopicsRes[topicName]
		if !exists {
			return fmt.Errorf("requested create topic not part of the create response")
		}
		if topicRes.Err != nil {
			return fmt.Errorf("failed to create topic: %w", topicRes.Err)
		}
	}

	// TODO: Check if Kafka topic config matches
	return nil
}
