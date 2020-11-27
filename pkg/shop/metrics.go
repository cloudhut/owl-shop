package shop

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	EventTypeAddressCreated = "ADDRESS_CREATED"

	EventTypeCustomerCreated  = "CUSTOMER_CREATED"
	EventTypeCustomerModified = "CUSTOMER_MODIFIED"
	EventTypeCustomerDeleted  = "CUSTOMER_DELETED"
	EventTypeCustomerConsumed = "CUSTOMER_CONSUMED"

	EventTypeOrderCreated = "ORDER_CREATED"

	EventTypeFrontendEventCreated = "FRONTEND_EVENT_CREATED"
)

var (
	promNamespace = "owl_shop"

	pageImpressionsSimulated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: promNamespace,
		Name:      "simulated_impressions_total",
		Help:      "The number of page impressions simulated",
	})

	kafkaMessagesProducedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: promNamespace,
		Name:      "kafka_messages_produced_total",
		Help:      "The number of Kafka messages produced to a Kafka topic",
	}, []string{"event_type"})
	kafkaMessagesConsumedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: promNamespace,
		Name:      "kafka_messages_consumed_total",
		Help:      "The number of Kafka messages consumed",
	}, []string{"event_type"})
)
