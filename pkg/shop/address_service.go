package shop

import (
	"context"
	"encoding/json"
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

type AddressService struct {
	cfg            Config
	logger         *zap.Logger
	kafkaSvc       *kafka.Service
	consumerClient *kgo.Client

	bufferSize       int
	recentCustomerMu sync.RWMutex
	recentCustomers  []fake.Customer

	clientID  string
	topicName string
}

func NewAddressService(cfg Config, logger *zap.Logger) (*AddressService, error) {
	clientID := cfg.GlobalPrefix + "address-service"
	cfg.Kafka.ClientID = clientID
	kafkaSvc, err := kafka.NewService(cfg.Kafka, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka service: %w", err)
	}

	consumerClient, err := kafkaSvc.NewKafkaClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer client: %w", err)
	}

	// This slice is used to keep some customers in the buffer so that we can produce addresses for these customers
	bufferSize := 500
	recentCustomers := make([]fake.Customer, 0, bufferSize)

	return &AddressService{
		cfg:            cfg,
		logger:         logger.With(zap.String("service", "address_service")),
		kafkaSvc:       kafkaSvc,
		consumerClient: consumerClient,

		bufferSize:       bufferSize,
		recentCustomerMu: sync.RWMutex{},
		recentCustomers:  recentCustomers,

		clientID:  clientID,
		topicName: cfg.GlobalPrefix + "addresses",
	}, nil
}

func (svc *AddressService) Initialize(ctx context.Context) error {
	svc.logger.Info("initializing address service")

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

	return nil
}

func (svc *AddressService) Start() {
	svc.consumerClient.AssignGroup(svc.clientID,
		kgo.GroupTopics(svc.cfg.GlobalPrefix+"customers"),
		kgo.AutoCommitInterval(500*time.Millisecond))
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		fetches := svc.consumerClient.PollFetches(ctx)
		cancel()
		errors := fetches.Errors()
		if errors != nil {
			svc.logger.Warn("failed to poll fetches", zap.Error(errors[0].Err))
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			rec := iter.Next()
			kafkaMessagesConsumedTotal.With(map[string]string{"event_type": EventTypeCustomerConsumed}).Inc()

			if rec.Value == nil {
				continue
			}
			customer := fake.Customer{}
			err := json.Unmarshal(rec.Value, &customer)
			if err != nil {
				// Skip message
				svc.logger.Warn("failed to deserialize customer", zap.Error(err))
				continue
			}
			svc.recentCustomerMu.Lock()
			if len(svc.recentCustomers) < svc.bufferSize {
				svc.recentCustomers = append(svc.recentCustomers, customer)
			}
			svc.recentCustomerMu.Unlock()
		}
	}
}

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
	serialized, err := kafka.SerializeJson(address)
	if err != nil {
		return fmt.Errorf("failed to serialize customer struct: %w", err)
	}

	rec := kgo.Record{
		Key:       []byte(address.ID),
		Value:     serialized,
		Headers:   []kgo.RecordHeader{{Key: "revision", Value: []byte("0")}},
		Timestamp: time.Now(),
		Topic:     svc.topicName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = svc.kafkaSvc.KafkaClient.Produce(ctx, &rec, func(rec *kgo.Record, err error) {
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

// createKafkaTopic tries to create the Kafka topic
func (svc *AddressService) createKafkaTopic(ctx context.Context) error {
	cleanupPolicy := "compact"
	req := kmsg.CreateTopicsRequest{
		Topics: []kmsg.CreateTopicsRequestTopic{
			{
				Topic:             svc.topicName,
				NumPartitions:     6,
				ReplicationFactor: svc.cfg.Kafka.TopicReplicationFactor,
				Configs: []kmsg.CreateTopicsRequestTopicConfig{
					{"cleanup.policy", &cleanupPolicy},
				},
			},
		},
		TimeoutMillis: 15 * 1000,
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
