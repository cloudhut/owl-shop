package shop

import (
	"fmt"
	"time"
)

type ConfigTrafficInterval struct {
	// Number of requests per configured interval duration
	Rate int `yaml:"rate"`

	// Duration in which all requests shall be issued
	Duration time.Duration `yaml:"duration"`
}

func (c *ConfigTrafficInterval) Validate() error {
	if c.Rate <= 0 {
		return fmt.Errorf("interval rate must be higher than 0")
	}

	return nil
}

func (c *ConfigTrafficInterval) SetDefaults() {
	c.Rate = 5
	c.Duration = 10 * time.Second
}
