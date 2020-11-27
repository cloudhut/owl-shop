package shop

import (
	"context"
	"fmt"
	"github.com/cloudhut/owl-shop/pkg/fake"
	"github.com/cloudhut/owl-shop/pkg/kafka"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"go.uber.org/zap"
	"sync"
	"time"
)

type FrontendService struct {
	cfg      Config
	logger   *zap.Logger
	kafkaSvc *kafka.Service

	clientID  string
	topicName string
}

func NewFrontendService(cfg Config, logger *zap.Logger) (*FrontendService, error) {
	cfg.Kafka.ClientID = cfg.GlobalPrefix + "frontend-service"
	kafkaSvc, err := kafka.NewService(cfg.Kafka, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka service: %w", err)
	}

	return &FrontendService{
		cfg:      cfg,
		logger:   logger.With(zap.String("service", "frontend_service")),
		kafkaSvc: kafkaSvc,

		topicName: cfg.GlobalPrefix + "frontend-events",
	}, nil
}

func (svc *FrontendService) Initialize(ctx context.Context) error {
	svc.logger.Info("initializing customer service")

	// 1. Test kafka connectivity
	metadata, err := svc.kafkaSvc.GetMetadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to get metadata to test kafka connectivity: %w", err)
	}

	// 2. Ensure that Kafka topic exists
	isTopicExistent := false
	for _, topic := range metadata.Topics {
		if topic.Topic == svc.topicName {
			isTopicExistent = true
			break
		}
	}
	if !isTopicExistent {
		err := svc.createKafkaTopic(ctx)
		if err != nil {
			return fmt.Errorf("failed to create kafka topic '%v': %w", svc.topicName, err)
		}
		svc.logger.Info("successfully created Kafka topic", zap.String("topic_name", svc.topicName))
	}

	svc.logger.Info("successfully initialized customer service")

	return nil
}

func (svc *FrontendService) CreateFrontendEvent() {
	event := fake.NewFrontendEvent()
	err := svc.produceFrontendEvent(event)
	if err != nil {
		svc.logger.Warn("failed to produce address", zap.Error(err))
		return
	}
	kafkaMessagesProducedTotal.With(map[string]string{"event_type": EventTypeFrontendEventCreated}).Inc()
}

func (svc *FrontendService) produceFrontendEvent(event fake.FrontendEvent) error {
	serialized, err := kafka.SerializeJson(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event struct: %w", err)
	}

	rec := kgo.Record{
		Key:       nil,
		Value:     serialized,
		Headers:   nil,
		Timestamp: time.Now(),
		Topic:     svc.topicName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	err = svc.kafkaSvc.KafkaClient.Produce(ctx, &rec, func(rec *kgo.Record, err error) {
		defer wg.Done()
		if err != nil {
			svc.logger.Error("failed to produce record",
				zap.String("topic_name", rec.Topic),
				zap.Error(err),
			)
			return
		}
	})
	if err != nil {
		return fmt.Errorf("failed to produce: %w", err)
	}
	wg.Wait()

	return nil
}

// createKafkaTopic tries to create the Kafka topic
func (svc *FrontendService) createKafkaTopic(ctx context.Context) error {
	cleanupPolicy := "delete"
	retentionBytes := "1000000000" // 1GB
	segmentBytes := "1000000"      // 100MB
	req := kmsg.CreateTopicsRequest{
		Topics: []kmsg.CreateTopicsRequestTopic{
			{
				Topic:             svc.topicName,
				NumPartitions:     6,
				ReplicationFactor: 3,
				Configs: []kmsg.CreateTopicsRequestTopicConfig{
					{"cleanup.policy", &cleanupPolicy},
					{"retention.bytes", &retentionBytes},
					{"segment.bytes", &segmentBytes},
				},
			},
		},
		TimeoutMillis: 60 * 1000,
	}
	res, err := req.RequestWith(ctx, svc.kafkaSvc.KafkaClient)
	if err != nil {
		return fmt.Errorf("create topics request failed: %w", err)
	}

	// Check for inner kafka errors
	if len(res.Topics) != 1 {
		return fmt.Errorf("unexpected topic response count from create topics request. Expected count '1', actual returned count: '%v'", len(res.Topics))
	}

	err = kerr.ErrorForCode(res.Topics[0].ErrorCode)
	if err != nil {
		return fmt.Errorf("create topics request failed. Inner kafka error: %w", err)
	}

	return nil
}
