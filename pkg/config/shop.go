package config

import "time"

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
}

func (c *Shop) SetDefaults() {
	c.GlobalPrefix = "owlshop-"
}

func (c *Shop) Validate() error {
	return nil
}
