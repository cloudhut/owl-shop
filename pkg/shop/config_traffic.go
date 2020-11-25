package shop

import "fmt"

type TrafficPattern = string

const (
	TrafficPatternConstant TrafficPattern = "constant"
)

type ConfigTraffic struct {
	Pattern  string                `yaml:"pattern"`
	Interval ConfigTrafficInterval `yaml:"interval"`
}

func (c *ConfigTraffic) SetDefaults() {
	c.Pattern = TrafficPatternConstant
	c.Interval.SetDefaults()
}

func (c *ConfigTraffic) Validate() error {
	if c.Pattern != TrafficPatternConstant {
		return fmt.Errorf("given traffic pattern '%v' is invalid", c.Pattern)
	}

	err := c.Interval.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate interval config: %w", err)
	}

	return nil
}
