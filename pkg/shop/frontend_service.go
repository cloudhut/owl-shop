package shop

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/cloudhut/owl-shop/pkg/fake"
	"github.com/cloudhut/owl-shop/pkg/kafka"
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
	err := svc.kafkaSvc.ReconcileTopic(
		ctx,
		svc.topicName,
		svc.cfg.Kafka.TopicPartitionCount,
		svc.cfg.Kafka.TopicReplicationFactor,
		map[string]*string{
			"cleanup.policy":  kadm.StringPtr("delete"),
			"retention.bytes": kadm.StringPtr("3221225472"), // 3GiB
		},
	)
	if err != nil {
		return fmt.Errorf("failed to reconcile topic: %w", err)
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

	svc.kafkaSvc.KafkaClient.Produce(context.Background(), &rec, func(rec *kgo.Record, err error) {
		if err != nil {
			svc.logger.Error("failed to produce record",
				zap.String("topic_name", rec.Topic),
				zap.Error(err),
			)
			return
		}
	})

	return nil
}
