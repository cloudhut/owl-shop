package shop

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/cloudhut/owl-shop/pkg/config"
	"github.com/cloudhut/owl-shop/pkg/fake"
	"github.com/cloudhut/owl-shop/pkg/kafka"
)

// AddressService consumes the customers topic to collect customer ID and name
// and then produces fake addresses for that customer.
type AddressService struct {
	cfg          config.Shop
	logger       *zap.Logger
	kafkaFactory *kafka.Factory

	metaClient     *kgo.Client
	consumerClient *kgo.Client

	bufferSize       int
	recentCustomerMu sync.RWMutex
	recentCustomers  []fake.Customer

	clientID  string
	topicName string
}

// NewAddressService creates the service that publishes addresses to the
// address topic.
func NewAddressService(
	cfg config.Shop,
	logger *zap.Logger,
	kafkaFactory *kafka.Factory,
) (*AddressService, error) {
	clientID := cfg.GlobalPrefix + "address-service"
	topicName := cfg.GlobalPrefix + "addresses"
	consumerClient, err := kafkaFactory.NewKafkaClient(
		clientID,
		kgo.ConsumeTopics(topicName),
		kgo.ConsumerGroup(clientID),
		kgo.AutoCommitInterval(500*time.Millisecond),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer client: %w", err)
	}

	metaClient, err := kafkaFactory.NewKafkaClient(clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to create meta client: %w", err)
	}

	// This slice is used to keep some customers in the buffer so that we can produce addresses for these customers
	bufferSize := 500
	recentCustomers := make([]fake.Customer, 0, bufferSize)

	return &AddressService{
		cfg:          cfg,
		logger:       logger.With(zap.String("service", "address_service")),
		kafkaFactory: kafkaFactory,

		consumerClient: consumerClient,
		metaClient:     metaClient,

		bufferSize:       bufferSize,
		recentCustomerMu: sync.RWMutex{},
		recentCustomers:  recentCustomers,

		clientID:  clientID,
		topicName: cfg.GlobalPrefix + "addresses",
	}, nil
}

// Initialize address service.
func (svc *AddressService) Initialize(ctx context.Context) error {
	svc.logger.Info("initializing address service")

	err := kafka.ReconcileTopic(ctx,
		svc.metaClient,
		svc.topicName,
		svc.cfg.TopicPartitionCount,
		svc.cfg.TopicReplicationFactor,
		map[string]*string{
			"cleanup.policy": kadm.StringPtr("compact"),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to reconcile topic: %w", err)
	}

	return nil
}

// Start consuming messages from customers topic that are required
// to produce address records.
func (svc *AddressService) Start() {
	for {
		fetches := svc.consumerClient.PollFetches(context.Background())

		if fetches.IsClientClosed() {
			svc.logger.Warn("client closed")
			return
		}

		fetches.EachError(func(topic string, partition int32, err error) {
			svc.logger.Error("failed to poll fetches",
				zap.String("topic", topic),
				zap.Int32("partition", partition),
				zap.Error(err))
		})

		fetches.EachRecord(func(rec *kgo.Record) {
			kafkaMessagesConsumedTotal.
				With(map[string]string{"event_type": EventTypeCustomerConsumed}).
				Inc()

			if rec.Value == nil {
				return
			}

			customer := fake.Customer{}
			err := json.Unmarshal(rec.Value, &customer)
			if err != nil {
				// Skip message
				svc.logger.Warn("failed to deserialize customer", zap.Error(err))
				return
			}
			svc.recentCustomerMu.Lock()
			if len(svc.recentCustomers) < svc.bufferSize {
				svc.recentCustomers = append(svc.recentCustomers, customer)
			}
			svc.recentCustomerMu.Unlock()
		})
	}
}

// CreateAddress produces a new fake address record and produces that record
// to the address topic.
func (svc *AddressService) CreateAddress() {
	customer, err := svc.popCustomerFromBuffer()
	if err != nil {
		svc.logger.Debug("failed to pop customer from buffer", zap.Error(err))
		return
	}
	address := fake.NewAddress(customer)
	err = svc.produceAddress(address)
	if err != nil {
		svc.logger.Warn("failed to produce address", zap.Error(err))
		return
	}
	kafkaMessagesProducedTotal.With(map[string]string{"event_type": EventTypeAddressCreated}).Inc()
}

func (svc *AddressService) produceAddress(address fake.Address) error {
	serialized, err := json.Marshal(address)
	if err != nil {
		return fmt.Errorf("failed to serialize customer struct: %w", err)
	}

	rec := kgo.Record{
		Key:     []byte(address.ID),
		Value:   serialized,
		Headers: []kgo.RecordHeader{{Key: "revision", Value: []byte("0")}},
		Topic:   svc.topicName,
	}

	svc.metaClient.Produce(context.Background(), &rec, func(rec *kgo.Record, err error) {
		if err == nil {
			return
		}
		svc.logger.Error("failed to produce record",
			zap.String("topic_name", rec.Topic),
			zap.Error(err),
		)
	})
	return nil
}

func (svc *AddressService) popCustomerFromBuffer() (fake.Customer, error) {
	svc.recentCustomerMu.Lock()
	defer svc.recentCustomerMu.Unlock()

	if len(svc.recentCustomers) == 0 {
		// No customers in buffer yet
		return fake.Customer{}, fmt.Errorf("buffer is empty")
	}
	customer := svc.recentCustomers[0]
	svc.recentCustomers = svc.recentCustomers[1:]

	return customer, nil
}
