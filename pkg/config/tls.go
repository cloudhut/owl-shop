package config

import (
	"crypto/tls"

	"github.com/twmb/tlscfg"
)

// TLS configuration for either the HTTP protocol or Kafka API.
type TLS struct {
	Enabled               bool   `yaml:"enabled"`
	CaFilepath            string `yaml:"caFilepath"`
	CertFilepath          string `yaml:"certFilepath"`
	KeyFilepath           string `yaml:"keyFilepath"`
	InsecureSkipTLSVerify bool   `yaml:"insecureSkipTlsVerify"`
}

// TLSConfig builds a tls.Config based on the configuration.
func (c *TLS) TLSConfig() (*tls.Config, error) {
	return tlscfg.New(
		tlscfg.MaybeWithDiskCA(c.CaFilepath, tlscfg.ForClient),
		tlscfg.MaybeWithDiskKeyPair(c.CertFilepath, c.KeyFilepath),
		tlscfg.WithOverride(func(config *tls.Config) error {
			if c.InsecureSkipTLSVerify {
				config.InsecureSkipVerify = true
			}
			return nil
		}),
	)
}
