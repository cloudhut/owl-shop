package config

// SchemaRegistry is the configuration for the schema registry.
type SchemaRegistry struct {
	Address   string        `yaml:"address"`
	BasicAuth HTTPBasicAuth `yaml:"basicAuth"`
	TLS       TLS           `yaml:"tls"`
}

// HTTPBasicAuth for authentication via HTTP.
type HTTPBasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Validate SchemaRegistry config.
func (c *SchemaRegistry) Validate() error {
	if c.Address == "" {
		return nil
	}

	return nil
}

// SetDefaults for SchemaRegistry config
func (c *SchemaRegistry) SetDefaults() {
	// For HTTP protocols this setting doesn't make as much sense
	c.TLS.Enabled = true
}
