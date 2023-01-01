package shop

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/cloudhut/owl-shop/pkg/config"
	"github.com/cloudhut/owl-shop/pkg/fake"
	"github.com/cloudhut/owl-shop/pkg/kafka"
)

type FrontendService struct {
	cfg    config.Shop
	logger *zap.Logger

	kafkaFactory *kafka.Factory
	metaClient   *kgo.Client

	topicName string
}

func NewFrontendService(
	cfg config.Shop,
	logger *zap.Logger,
	kafkaFactory *kafka.Factory,
) (*FrontendService, error) {
	clientID := cfg.GlobalPrefix + "frontend-service"
	metaClient, err := kafkaFactory.NewKafkaClient(clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	return &FrontendService{
		cfg:    cfg,
		logger: logger.With(zap.String("service", "frontend_service")),

		kafkaFactory: kafkaFactory,
		metaClient:   metaClient,

		topicName: cfg.GlobalPrefix + "frontend-events",
	}, nil
}

func (svc *FrontendService) Initialize(ctx context.Context) error {
	svc.logger.Info("initializing customer service")
	err := kafka.ReconcileTopic(
		ctx,
		svc.metaClient,
		svc.topicName,
		svc.cfg.TopicPartitionCount,
		svc.cfg.TopicReplicationFactor,
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
	serialized, err := json.Marshal(event)
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

	svc.metaClient.Produce(context.Background(), &rec, func(rec *kgo.Record, err error) {
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
