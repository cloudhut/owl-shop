package shop

import (
	"flag"
	"fmt"
	"github.com/cloudhut/owl-shop/pkg/kafka"
)

type Config struct {
	Kafka kafka.Config `yaml:"kafka"`

	Traffic ConfigTraffic `yaml:"traffic"`
	// Prefix for all topic names, consumer group names, client ids etc.
	GlobalPrefix string `yaml:"globalPrefix"`
}

func (c *Config) SetDefaults() {
	c.Kafka.SetDefaults()
	c.GlobalPrefix = "owlshop-"
}

func (c *Config) Validate() error {
	err := c.Kafka.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate kafka cfg: %w", err)
	}

	return nil
}

// RegisterFlags for all (sub)configs
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Kafka.RegisterFlags(f)
}
