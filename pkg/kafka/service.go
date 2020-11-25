package kafka

import (
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

// Service acts as interface to interact with the Kafka Cluster
type Service struct {
	Config      Config
	Logger      *zap.Logger
	KafkaClient *kgo.Client
}

// NewService creates a new Kafka service and immediately checks connectivity to all components. If any of these external
// dependencies fail an error wil be returned.
func NewService(cfg Config, logger *zap.Logger) (*Service, error) {
	kgoOpts, err := NewKgoConfig(&cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create a valid kafka client config: %w", err)
	}

	kafkaClient, err := kgo.NewClient(kgoOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	return &Service{
		Config:      cfg,
		Logger:      logger,
		KafkaClient: kafkaClient,
	}, nil
}

func (s *Service) NewKafkaClient() (*kgo.Client, error) {
	kgoOpts, err := NewKgoConfig(&s.Config, s.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create a valid kafka client config: %w", err)
	}

	kafkaClient, err := kgo.NewClient(kgoOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	return kafkaClient, nil
}
