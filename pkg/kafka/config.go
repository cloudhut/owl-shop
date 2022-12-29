package kafka

import (
	"flag"
	"fmt"
)

type Config struct {
	// General
	Brokers  []string `yaml:"brokers"`
	ClientID string   `yaml:"clientId"`

	TLS  TLSConfig  `yaml:"tls"`
	SASL SASLConfig `yaml:"sasl"`

	TopicReplicationFactor int16 `yaml:"topicReplicationFactor"`
	TopicPartitionCount    int32 `yaml:"topicPartitionCount"`
}

// RegisterFlags for all sensitive Kafka SASL configs.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.TLS.RegisterFlags(f)
	c.SASL.RegisterFlags(f)
}

func (c *Config) Validate() error {
	if len(c.Brokers) == 0 {
		return fmt.Errorf("you must configure at least one broker to connect to")
	}

	err := c.SASL.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate SASL config: %w", err)
	}

	return nil
}

// SetDefaults for Kafka config
func (c *Config) SetDefaults() {
	c.ClientID = "owl-shop"
	c.TopicReplicationFactor = 3
	c.TopicPartitionCount = 6

	c.SASL.SetDefaults()
}
