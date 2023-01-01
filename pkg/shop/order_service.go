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
	"google.golang.org/protobuf/proto"

	"github.com/cloudhut/owl-shop/pkg/config"
	"github.com/cloudhut/owl-shop/pkg/fake"
	"github.com/cloudhut/owl-shop/pkg/kafka"
)

type OrderService struct {
	cfg    config.Shop
	logger *zap.Logger

	kafkaFactory   *kafka.Factory
	consumerClient *kgo.Client
	metaClient     *kgo.Client

	bufferSize        int
	recentCustomersMu sync.RWMutex
	recentCustomers   []fake.Customer

	topicName         string
	topicNameProtobuf string
}

func NewOrderService(
	cfg config.Shop,
	logger *zap.Logger,
	kafkaFactory *kafka.Factory,
) (*OrderService, error) {
	clientID := cfg.GlobalPrefix + "order-service"

	metaClient, err := kafkaFactory.NewKafkaClient(clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka service: %w", err)
	}

	consumerClient, err := kafkaFactory.NewKafkaClient(
		clientID,
		kgo.ConsumerGroup(clientID),
		kgo.ConsumeTopics(cfg.GlobalPrefix+"customers"),
		kgo.AutoCommitInterval(500*time.Millisecond),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer client: %w", err)
	}

	// This slice is used to keep some customers in the buffer so that they can be modified or deleted
	bufferSize := 500
	recentCustomers := make([]fake.Customer, 0, bufferSize)

	return &OrderService{
		cfg:    cfg,
		logger: logger.With(zap.String("service", "order_service")),

		kafkaFactory:   kafkaFactory,
		consumerClient: consumerClient,
		metaClient:     metaClient,

		bufferSize:        bufferSize,
		recentCustomersMu: sync.RWMutex{},
		recentCustomers:   recentCustomers,

		topicName:         cfg.GlobalPrefix + "orders",
		topicNameProtobuf: cfg.GlobalPrefix + "orders" + "-protobuf",
	}, nil
}

func (svc *OrderService) Start() {
	for {
		fetches := svc.consumerClient.PollFetches(context.Background())

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
			svc.recentCustomersMu.Lock()
			if len(svc.recentCustomers) < svc.bufferSize {
				svc.recentCustomers = append(svc.recentCustomers, customer)
			}
			svc.recentCustomersMu.Unlock()
		}
	}
}

func (svc *OrderService) Initialize(ctx context.Context) error {
	svc.logger.Info("initializing order service")

	topicCfg := map[string]*string{
		"cleanup.policy": kadm.StringPtr("compact"),
	}
	err := kafka.ReconcileTopic(ctx,
		svc.metaClient,
		svc.topicName,
		svc.cfg.TopicPartitionCount,
		svc.cfg.TopicReplicationFactor,
		topicCfg,
	)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	err = kafka.ReconcileTopic(ctx,
		svc.metaClient,
		svc.topicNameProtobuf,
		svc.cfg.TopicPartitionCount,
		svc.cfg.TopicReplicationFactor,
		topicCfg,
	)
	if err != nil {
		return fmt.Errorf("failed to create protobuf topic: %w", err)
	}

	return nil
}

func (svc *OrderService) CreateOrder() {
	customer, err := svc.popCustomerFromBuffer()
	if err != nil {
		svc.logger.Debug("failed to pop customer from buffer", zap.Error(err))
		return
	}
	order := fake.NewOrder(customer)

	err = svc.produceOrderJSON(order)
	if err != nil {
		svc.logger.Warn("failed to produce order (json)", zap.Error(err))
		return
	}
	err = svc.produceOrderProtobuf(order)
	if err != nil {
		svc.logger.Warn("failed to produce order (protobuf)", zap.Error(err))
		return
	}
	kafkaMessagesProducedTotal.With(map[string]string{"event_type": EventTypeOrderCreated}).Add(2)
	return
}

func (svc *OrderService) produceOrderJSON(order fake.Order) error {
	serialized, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to serialize customer struct: %w", err)
	}

	rec := kgo.Record{
		Key:       []byte(order.ID),
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

	return nil
}

func (svc *OrderService) produceOrderProtobuf(order fake.Order) error {
	pbOrder := order.Protobuf()
	serialized, err := proto.Marshal(pbOrder)
	if err != nil {
		return fmt.Errorf("failed to serialize customer struct: %w", err)
	}

	rec := kgo.Record{
		Key:   []byte(order.ID),
		Value: serialized,
		Headers: []kgo.RecordHeader{
			{Key: "revision", Value: []byte("0")},
			{Key: "proto_message_type", Value: []byte("Order")},
		},
		Timestamp: time.Now(),
		Topic:     svc.topicNameProtobuf,
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

func (svc *OrderService) popCustomerFromBuffer() (fake.Customer, error) {
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
