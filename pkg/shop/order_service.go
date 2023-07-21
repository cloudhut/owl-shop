package shop

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/cloudhut/owl-shop/pkg/config"
	"github.com/cloudhut/owl-shop/pkg/fake"
	"github.com/cloudhut/owl-shop/pkg/kafka"
	shoppb "github.com/cloudhut/owl-shop/pkg/protogen/shop/v1"
	embedproto "github.com/cloudhut/owl-shop/proto"
)

var (
	//go:embed customer_v1.proto
	customerV1Proto string
)

// OrderService is the service that is in charge of handling incoming orders.
// When a new customer order is received this service will produce a message
// on the order topics in different formats (JSON and Protobuf).
// Because orders belong to a customer, this service also consumes the customers
// topic
type OrderService struct {
	cfg    config.Shop
	logger *zap.Logger

	kafkaFactory   *kafka.Factory
	consumerClient *kgo.Client
	metaClient     *kgo.Client
	srClient       *sr.Client

	bufferSize        int
	recentCustomersMu sync.RWMutex
	recentCustomers   []fake.Customer

	topicName              string
	topicNameProtobufPlain string
	topicNameProtobufSr    string

	protobufSerde sr.Serde
}

// NewOrderService creates a new OrderService. All dependencies are passed into here.
// The schema registry client is optional and may be nil.
func NewOrderService(
	cfg config.Shop,
	logger *zap.Logger,
	kafkaFactory *kafka.Factory,
	srClient *sr.Client,
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
		srClient:       srClient,

		bufferSize:        bufferSize,
		recentCustomersMu: sync.RWMutex{},
		recentCustomers:   recentCustomers,

		topicName:              cfg.GlobalPrefix + "orders",
		topicNameProtobufPlain: cfg.GlobalPrefix + "orders" + "-protobuf-plain",
		topicNameProtobufSr:    cfg.GlobalPrefix + "orders" + "-protobuf-sr",

		protobufSerde: sr.Serde{}, // Has to be registered after creating the schema
	}, nil
}

// Start starts polling for new messages on the customers topic.
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

// Initialize order service by reconciling all order topics.
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
		svc.topicNameProtobufPlain,
		svc.cfg.TopicPartitionCount,
		svc.cfg.TopicReplicationFactor,
		topicCfg,
	)
	if err != nil {
		return fmt.Errorf("failed to create protobuf plain topic: %w", err)
	}

	if svc.srClient != nil {
		// Protobuf topic
		if err := kafka.ReconcileTopic(ctx,
			svc.metaClient,
			svc.topicNameProtobufSr,
			svc.cfg.TopicPartitionCount,
			svc.cfg.TopicReplicationFactor,
			topicCfg,
		); err != nil {
			return fmt.Errorf("failed to create protobuf sr topic: %w", err)
		}

		orderSchemaID, err := svc.registerProtobufSchema(ctx)
		if err != nil {
			return fmt.Errorf("failed to register protobuf schemas in schema registry: %w", err)
		}

		svc.protobufSerde.Register(
			orderSchemaID,
			&shoppb.Order{},
			sr.EncodeFn(func(v any) ([]byte, error) {
				return proto.Marshal(v.(*shoppb.Order))
			}),
			sr.Index(0),
		)
	}

	return nil
}

// registerProtobufSchema registers the used protobuf schemas in the schema registry, so that
// serialized messages can be deserialized by other tools like Redpanda Console or CLIs.
// If successful, it returns the schema id.
func (svc *OrderService) registerProtobufSchema(ctx context.Context) (int, error) {
	// Register dependency schemas first, then main schema with references
	customerProtoSubject := "shop/v1/customer.proto"

	// This registers an older proto version first, so that we simulate
	// a schema evolution as well.
	_, err := svc.srClient.CreateSchema(
		ctx,
		customerProtoSubject,
		sr.Schema{
			Schema: customerV1Proto,
			Type:   sr.TypeProtobuf,
		},
	)
	if err != nil {
		return -1, fmt.Errorf("failed to register customer schema: %w", err)
	}

	_, err = svc.srClient.CreateSchema(
		ctx,
		customerProtoSubject,
		sr.Schema{
			Schema: embedproto.Customer,
			Type:   sr.TypeProtobuf,
		},
	)
	if err != nil {
		return -1, fmt.Errorf("failed to register customer schema: %w", err)
	}

	addressProtoSubject := "shop/v1/address.proto"
	_, err = svc.srClient.CreateSchema(
		ctx,
		addressProtoSubject,
		sr.Schema{
			Schema: embedproto.Address,
			Type:   sr.TypeProtobuf,
		},
	)
	if err != nil {
		return -1, fmt.Errorf("failed to register address schema: %w", err)
	}

	orderSchema, err := svc.srClient.CreateSchema(
		ctx,
		svc.topicNameProtobufSr+"-value",
		sr.Schema{
			Schema: embedproto.Order,
			Type:   sr.TypeProtobuf,
			References: []sr.SchemaReference{
				{
					Name:    customerProtoSubject,
					Subject: customerProtoSubject,
					Version: 2,
				},
				{
					Name:    addressProtoSubject,
					Subject: addressProtoSubject,
					Version: 1,
				},
			},
		},
	)
	if err != nil {
		return -1, fmt.Errorf("failed to register order schema: %w", err)
	}

	return orderSchema.ID, nil
}

// CreateOrder creates a new fake order message. It pops a previously produced
// fake customer from the in-memory cache so that an existing customer can be
// referenced in the order message.
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
	err = svc.produceOrderPlainProtobuf(order)
	if err != nil {
		svc.logger.Warn("failed to produce order (protobuf)", zap.Error(err))
		return
	}
	kafkaMessagesProducedTotal.With(map[string]string{"event_type": EventTypeOrderCreated}).Add(2)

	if svc.srClient != nil {
		err = svc.produceOrderSrProtobuf(order)
		if err != nil {
			svc.logger.Warn("failed to produce order (protobuf)", zap.Error(err))
			return
		}
		kafkaMessagesProducedTotal.With(map[string]string{"event_type": EventTypeOrderCreated}).Add(1)
	}

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

func (svc *OrderService) produceOrderPlainProtobuf(order fake.Order) error {
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
		Topic:     svc.topicNameProtobufPlain,
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

// produceOrderSrProtobuf produces a protobuf message with schema registry encoding.
func (svc *OrderService) produceOrderSrProtobuf(order fake.Order) error {
	pbOrder := order.Protobuf()
	serialized, err := svc.protobufSerde.Encode(pbOrder)
	if err != nil {
		return fmt.Errorf("failed to encode porotobuf order: %w", err)
	}

	rec := kgo.Record{
		Key:   []byte(order.ID),
		Value: serialized,
		Headers: []kgo.RecordHeader{
			{Key: "revision", Value: []byte("0")},
			{Key: "proto_message_type", Value: []byte("Order")},
		},
		Timestamp: time.Now(),
		Topic:     svc.topicNameProtobufSr,
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
