package config

import (
	"fmt"
)

const (
	SASLMechanismPlain       = "PLAIN"
	SASLMechanismScramSHA256 = "SCRAM-SHA-256"
	SASLMechanismScramSHA512 = "SCRAM-SHA-512"
	SASLMechanismGSSAPI      = "GSSAPI"
	SASLMechanismOAuthBearer = "OAUTHBEARER"
)

// SASL config for Kafka client
type SASL struct {
	Enabled      bool             `yaml:"enabled"`
	Username     string           `yaml:"username"`
	Password     string           `yaml:"password"`
	Mechanism    string           `yaml:"mechanism"`
	GSSAPIConfig SASLGSSAPIConfig `yaml:"gssapi"`
}

// SetDefaults for SASL Config
func (c *SASL) SetDefaults() {
	c.Mechanism = SASLMechanismPlain
}

// Validate SASL config input
func (c *SASL) Validate() error {
	switch c.Mechanism {
	case SASLMechanismPlain, SASLMechanismScramSHA256, SASLMechanismScramSHA512, SASLMechanismGSSAPI:
		// Valid and supported
	case SASLMechanismOAuthBearer:
		return fmt.Errorf("sasl mechanism '%v' is valid but not yet supported. Please submit an issue if you need it", c.Mechanism)
	default:
		return fmt.Errorf("given sasl mechanism '%v' is invalid", c.Mechanism)
	}

	return nil
}
