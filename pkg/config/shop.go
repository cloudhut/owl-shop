package config

import (
	"fmt"
	"time"
)

// Shop is the configuration for the virtual shop that emits Kafka records
// upon simulated page impressions.
type Shop struct {
	// RequestRate is the number of requests per rate interval that shall
	// be simulated on the shop. Defaults to 2 requests at a default interval
	// of 1s.
	RequestRate int `yaml:"requestRate"`

	// RequestRateInterval is the interval in which {RequestRate} requests
	// shall be performed. Defaults to 1s.
	RequestRateInterval time.Duration `yaml:"interval"`

	// Prefix for all topic names, consumer group names, client ids etc.
	GlobalPrefix string `yaml:"globalPrefix"`

	// TopicReplicationFactor that shall be used for all Kafka topics.
	TopicReplicationFactor int16 `yaml:"topicReplicationFactor"`

	// TopicPartitionCount that shall be used for all Kafka topics.
	TopicPartitionCount int32 `yaml:"topicPartitionCount"`

	// Meta is the config for the meta service that creates additional
	// resources such as ACLs that are not required for generating
	// data, but may help to create a more production-like environment.
	Meta ShopMeta `yaml:"meta"`
}

// SetDefaults for shop config.
func (c *Shop) SetDefaults() {
	c.GlobalPrefix = "owlshop-"
	c.RequestRate = 2
	c.RequestRateInterval = time.Second
	c.TopicReplicationFactor = -1
	c.TopicPartitionCount = 1
	c.Meta.Enabled = true
}

// Validate shop configuration.
func (c *Shop) Validate() error {
	if c.RequestRate <= 0 {
		return fmt.Errorf("request rate must be a positive integer")
	}

	if c.RequestRateInterval == 0 {
		return fmt.Errorf("request rate must be a valid duration (e.g. '1s')")
	}

	if c.TopicReplicationFactor < -1 || c.TopicReplicationFactor == 0 {
		return fmt.Errorf("replication factor must be a positive integer or '-1' for using default replication factor")
	}

	if c.TopicPartitionCount < -1 || c.TopicPartitionCount == 0 {
		return fmt.Errorf("partition count must be a positive integer or '-1' for using the default partition count")
	}

	return nil
}
