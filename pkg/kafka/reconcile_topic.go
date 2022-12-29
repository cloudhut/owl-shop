package kafka

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// ReconcileTopic ensures a topic with the given parameters exist in the target cluster.
func (s *Service) ReconcileTopic(
	ctx context.Context,
	topicName string,
	partitions int32,
	replicationFactor int16,
	configs map[string]*string,
) error {
	topicDetails, err := s.KafkaAdminClient.ListTopics(ctx, topicName)
	if err != nil {
		return fmt.Errorf("failed to get metadata to check if topic exists: %w", err)
	}

	if !topicDetails.Has(topicName) {
		createTopicsRes, err := s.KafkaAdminClient.CreateTopics(
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
		s.Logger.Info("successfully created topic", zap.String("topic_name", topicRes.Topic))
	}

	// TODO: Check if Kafka topic config matches
	return nil
}
