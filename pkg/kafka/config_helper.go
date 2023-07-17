package kafka

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/jcmturner/gokrb5/v8/client"
	krbconfig "github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl"
	"github.com/twmb/franz-go/pkg/sasl/kerberos"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"github.com/twmb/franz-go/plugin/kzap"
	"github.com/twmb/tlscfg"
	"go.uber.org/zap"

	"github.com/cloudhut/owl-shop/pkg/config"
)

// NewKgoConfig creates a new Config for the Kafka Client as exposed by the franz-go library.
// If TLS certificates can't be read an error will be returned.
func NewKgoConfig(cfg *config.Kafka, logger *zap.Logger) ([]kgo.Opt, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.RecordDeliveryTimeout(10 * time.Second),
		kgo.WithLogger(kzap.New(logger)),
	}

	// Configure SASL
	if cfg.SASL.Enabled {
		// SASL Plain
		if cfg.SASL.Mechanism == "PLAIN" {
			mechanism := plain.Auth{
				User: cfg.SASL.Username,
				Pass: cfg.SASL.Password,
			}.AsMechanism()
			opts = append(opts, kgo.SASL(mechanism))
		}

		// SASL SCRAM
		if cfg.SASL.Mechanism == "SCRAM-SHA-256" || cfg.SASL.Mechanism == "SCRAM-SHA-512" {
			var mechanism sasl.Mechanism
			scramAuth := scram.Auth{
				User: cfg.SASL.Username,
				Pass: cfg.SASL.Password,
			}
			if cfg.SASL.Mechanism == "SCRAM-SHA-256" {
				mechanism = scramAuth.AsSha256Mechanism()
			}
			if cfg.SASL.Mechanism == "SCRAM-SHA-512" {
				mechanism = scramAuth.AsSha512Mechanism()
			}
			opts = append(opts, kgo.SASL(mechanism))
		}

		// Kerberos
		if cfg.SASL.Mechanism == "GSSAPI" {
			kerbCfg, err := krbconfig.Load(cfg.SASL.GSSAPIConfig.KerberosConfigPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create kerberos config from specified config filepath: %w", err)
			}
			var krbClient *client.Client
			switch cfg.SASL.GSSAPIConfig.AuthType {
			case "USER_AUTH:":
				krbClient = client.NewWithPassword(
					cfg.SASL.GSSAPIConfig.Username,
					cfg.SASL.GSSAPIConfig.Realm,
					cfg.SASL.GSSAPIConfig.Password,
					kerbCfg)
			case "KEYTAB_AUTH":
				ktb, err := keytab.Load(cfg.SASL.GSSAPIConfig.KeyTabPath)
				if err != nil {
					return nil, fmt.Errorf("failed to load keytab: %w", err)
				}
				krbClient = client.NewWithKeytab(
					cfg.SASL.GSSAPIConfig.Username,
					cfg.SASL.GSSAPIConfig.Realm,
					ktb,
					kerbCfg)
			}
			kerberosMechanism := kerberos.Auth{
				Client:           krbClient,
				Service:          cfg.SASL.GSSAPIConfig.ServiceName,
				PersistAfterAuth: true,
			}.AsMechanism()
			opts = append(opts, kgo.SASL(kerberosMechanism))
		}
	}

	// Configure TLS
	if cfg.TLS.Enabled {
		tlsCfg, err := tlscfg.New(
			tlscfg.MaybeWithDiskCA(cfg.TLS.CaFilepath, tlscfg.ForClient),
			tlscfg.MaybeWithDiskKeyPair(cfg.TLS.CertFilepath, cfg.TLS.KeyFilepath),
			tlscfg.WithSystemCertPool(),
			tlscfg.WithOverride(func(config *tls.Config) error {
				if cfg.TLS.InsecureSkipTLSVerify {
					config.InsecureSkipVerify = true
				}
				return nil
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create tls config: %w", err)
		}

		opts = append(opts, kgo.DialTLSConfig(tlsCfg))
	}

	return opts, nil
}
