package shop

import (
	"context"
	"fmt"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/cloudhut/owl-shop/pkg/fake"
	"github.com/cloudhut/owl-shop/pkg/kafka"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"go.uber.org/zap"
	"sync"
	"time"
)

type CustomerService struct {
	cfg      Config
	logger   *zap.Logger
	kafkaSvc *kafka.Service

	bufferSize        int
	recentCustomersMu sync.RWMutex
	recentCustomers   []fake.Customer

	topicName string
}

func NewCustomerService(cfg Config, logger *zap.Logger) (*CustomerService, error) {
	cfg.Kafka.ClientID = cfg.GlobalPrefix + "customer-service"
	kafkaSvc, err := kafka.NewService(cfg.Kafka, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka service: %w", err)
	}

	// This slice is used to keep some customers in the buffer so that they can be modified or deleted
	bufferSize := 500
	recentCustomers := make([]fake.Customer, 0, bufferSize)

	return &CustomerService{
		cfg:      cfg,
		logger:   logger.With(zap.String("service", "customer_service")),
		kafkaSvc: kafkaSvc,

		bufferSize:        bufferSize,
		recentCustomersMu: sync.RWMutex{},
		recentCustomers:   recentCustomers,

		topicName: cfg.GlobalPrefix + "customers",
	}, nil
}

func (svc *CustomerService) Initialize(ctx context.Context) error {
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
	return
}

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
	return
}

func (svc *CustomerService) DeleteCustomer() {
	customer, err := svc.popCustomerFromBuffer()
	if err != nil {
		svc.logger.Debug("failed to pop customer from buffer", zap.Error(err))
		return
	}

	svc.logger.Debug("deleted customer")

	err = svc.produceTombstone(customer.ID)
	if err != nil {
		svc.logger.Warn("failed to produce customer tombstone", zap.Error(err))
		return
	}
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

func (svc *CustomerService) produceTombstone(customerID string) error {
	rec := kgo.Record{
		Key:       []byte(customerID),
		Value:     nil,
		Timestamp: time.Now(),
		Topic:     svc.topicName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	err := svc.kafkaSvc.KafkaClient.Produce(ctx, &rec, func(rec *kgo.Record, err error) {
		defer wg.Done()
		if err != nil {
			svc.logger.Error("failed to produce tombstone record",
				zap.String("topic_name", rec.Topic),
				zap.Error(err),
			)
			return
		}
	})
	if err != nil {
		return fmt.Errorf("failed to produce tombstone: %w", err)
	}

	wg.Wait()
	return nil
}

func (svc *CustomerService) produceCustomer(customer fake.Customer) error {
	serialized, err := kafka.SerializeJson(customer)
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
func (svc *CustomerService) createKafkaTopic(ctx context.Context) error {
	cleanupPolicy := "compact"
	req := kmsg.CreateTopicsRequest{
		Topics: []kmsg.CreateTopicsRequestTopic{
			{
				Topic:             svc.topicName,
				NumPartitions:     6,
				ReplicationFactor: 3,
				Configs: []kmsg.CreateTopicsRequestTopicConfig{
					{"cleanup.policy", &cleanupPolicy},
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
