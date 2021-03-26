package kafka

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kmsg"
)

func (s *Service) GetMetadata(ctx context.Context) (*kmsg.MetadataResponse, error) {
	req := kmsg.MetadataRequest{
		Topics: nil,
	}

	return req.RequestWith(ctx, s.KafkaClient)
}

func (s *Service) GetTopicMetadata(ctx context.Context, topicName string) (*kmsg.MetadataResponseTopic, error) {
	req := kmsg.MetadataRequest{
		Topics: []kmsg.MetadataRequestTopic{
			{
				Topic: &topicName,
			},
		},
	}
	res, err := req.RequestWith(ctx, s.KafkaClient)
	if err != nil {
		return nil, fmt.Errorf("failed to request metadata: %w", err)
	}

	if len(res.Topics) == 0 {
		return nil, fmt.Errorf("expected exactly one topic result but got '%v'", err)
	}

	return &res.Topics[0], nil
}
