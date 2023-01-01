package kafka

import (
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/cloudhut/owl-shop/pkg/config"
)

// Factory allows requesters to create new preconfigured Kafka clients
// based on the given configuration.
type Factory struct {
	Config config.Kafka
	Logger *zap.Logger
}

// NewFactory creates a new Kafka factory.
func NewFactory(cfg config.Kafka, logger *zap.Logger) *Factory {
	return &Factory{
		Config: cfg,
		Logger: logger,
	}
}

// NewKafkaClient creates a new Kafka client with the same stored
// Kafka configuration.
func (s *Factory) NewKafkaClient(
	clientID string,
	additionalOpts ...kgo.Opt,
) (*kgo.Client, error) {
	kgoOpts, err := NewKgoConfig(&s.Config, s.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create a valid kafka client config: %w", err)
	}
	kgoOpts = append(kgoOpts, kgo.ClientID(clientID))
	kgoOpts = append(kgoOpts, additionalOpts...)

	kafkaClient, err := kgo.NewClient(kgoOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	return kafkaClient, nil
}
