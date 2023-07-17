package shop

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v5"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/cloudhut/owl-shop/pkg/config"
	"github.com/cloudhut/owl-shop/pkg/fake"
	"github.com/cloudhut/owl-shop/pkg/kafka"
)

// CustomerService emulates a service that produces a new Kafka record onto
// its topic every time a new user registers in the fake store. It produces to
// a compacted topic and regularly produces an updated version of an existing
// customer to simulate a change event. It also sends a tombstone for existing
// customers to simulate a delete request by a customer.
type CustomerService struct {
	cfg    config.Shop
	logger *zap.Logger

	kafkaFactory *kafka.Factory
	metaClient   *kgo.Client

	bufferSize        int
	recentCustomersMu sync.RWMutex
	recentCustomers   []fake.Customer

	topicName string
}

// NewCustomerService creates a new CustomerService.
func NewCustomerService(
	cfg config.Shop,
	logger *zap.Logger,
	kafkaFactory *kafka.Factory,
) (*CustomerService, error) {
	clientID := cfg.GlobalPrefix + "customer-service"
	metaClient, err := kafkaFactory.NewKafkaClient(clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	// This slice is used to keep some customers in the buffer so that they can be modified or deleted
	bufferSize := 500
	recentCustomers := make([]fake.Customer, 0, bufferSize)

	return &CustomerService{
		cfg:    cfg,
		logger: logger.With(zap.String("service", "customer_service")),

		kafkaFactory: kafkaFactory,
		metaClient:   metaClient,

		bufferSize:        bufferSize,
		recentCustomersMu: sync.RWMutex{},
		recentCustomers:   recentCustomers,

		topicName: cfg.GlobalPrefix + "customers",
	}, nil
}

// Initialize creates the customer topic with cleanup policy compact.
func (svc *CustomerService) Initialize(ctx context.Context) error {
	svc.logger.Info("initializing customer service")

	err := kafka.ReconcileTopic(
		ctx,
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

	svc.logger.Info("successfully initialized customer service")

	return nil
}

// CreateCustomer creates a fake customer struct and then produces the JSON serialized
// customer to the customer's topic.
func (svc *CustomerService) CreateCustomer() {
	customer := fake.NewCustomer()
	svc.recentCustomersMu.Lock()
	if len(svc.recentCustomers) < svc.bufferSize {
		svc.recentCustomers = append(svc.recentCustomers, customer)
	}
	svc.recentCustomersMu.Unlock()

	err := svc.produceCustomer(customer)
	if err != nil {
		svc.logger.Warn("failed to produce customer", zap.Error(err))
		return
	}
	kafkaMessagesProducedTotal.With(map[string]string{"event_type": EventTypeCustomerCreated}).Inc()
	return
}

// ModifyCustomer takes an existing customer from the cache, modifies the last name
// and sends the updated customer version to the customer's topic.
func (svc *CustomerService) ModifyCustomer() {
	customer, err := svc.popCustomerFromBuffer()
	if err != nil {
		svc.logger.Debug("failed to pop customer from buffer", zap.Error(err))
		return
	}

	customer.LastName = gofakeit.LastName()
	customer.Revision++
	svc.logger.Debug("modified customer")

	err = svc.produceCustomer(customer)
	if err != nil {
		svc.logger.Warn("failed to produce customer", zap.Error(err))
		return
	}
	kafkaMessagesProducedTotal.With(map[string]string{"event_type": EventTypeCustomerModified}).Inc()
	return
}

// DeleteCustomer sends a tombstone for an existing customer that was stored
// in the customer cache.
func (svc *CustomerService) DeleteCustomer() {
	customer, err := svc.popCustomerFromBuffer()
	if err != nil {
		svc.logger.Debug("failed to pop customer from buffer", zap.Error(err))
		return
	}

	svc.logger.Debug("deleted customer")

	svc.produceTombstone(customer.ID)
	if err != nil {
		svc.logger.Warn("failed to produce customer tombstone", zap.Error(err))
		return
	}
	kafkaMessagesProducedTotal.With(map[string]string{"event_type": EventTypeCustomerDeleted}).Inc()
}

func (svc *CustomerService) popCustomerFromBuffer() (fake.Customer, error) {
	svc.recentCustomersMu.Lock()
	defer svc.recentCustomersMu.Unlock()

	if len(svc.recentCustomers) == 0 {
		// No customers in buffer yet
		return fake.Customer{}, fmt.Errorf("buffer is empty")
	}
	customer := svc.recentCustomers[0]
	svc.recentCustomers = svc.recentCustomers[1:]

	return customer, nil
}

func (svc *CustomerService) produceTombstone(customerID string) {
	rec := kgo.Record{
		Key:       []byte(customerID),
		Value:     nil,
		Timestamp: time.Now(),
		Topic:     svc.topicName,
	}

	svc.metaClient.Produce(context.Background(), &rec, func(rec *kgo.Record, err error) {
		if err != nil {
			svc.logger.Error("failed to produce tombstone record",
				zap.String("topic_name", rec.Topic),
				zap.Error(err),
			)
			return
		}
	})
}

func (svc *CustomerService) produceCustomer(customer fake.Customer) error {
	serialized, err := json.Marshal(customer)
	if err != nil {
		return fmt.Errorf("failed to serialize customer struct: %w", err)
	}

	rec := kgo.Record{
		Key:       []byte(customer.ID),
		Value:     serialized,
		Headers:   []kgo.RecordHeader{{Key: "revision", Value: []byte("0")}},
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
	if err != nil {
		return fmt.Errorf("failed to produce: %w", err)
	}

	return nil
}
