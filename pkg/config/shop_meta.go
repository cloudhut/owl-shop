package config

// ShopMeta creates additional resources such as ACLs that are not required
// for generating data, but may be handy if you want to simulate a more
// production like environment, which is also used by other services.
type ShopMeta struct {
	Enabled bool `yaml:"enabled"`
}

func (c *ShopMeta) SetDefaults() {
	c.Enabled = true
}
