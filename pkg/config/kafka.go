package config

import (
	"fmt"
)

// Kafka API config for connecting to the target cluster.
type Kafka struct {
	Brokers []string `yaml:"brokers"`
	TLS     TLS      `yaml:"tls"`
	SASL    SASL     `yaml:"sasl"`
}

// Validate Kafka config.
func (c *Kafka) Validate() error {
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
func (c *Kafka) SetDefaults() {
	c.SASL.SetDefaults()
}
