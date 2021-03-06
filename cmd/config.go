package main

import (
	"flag"
	"fmt"
	"github.com/cloudhut/common/flagext"
	"github.com/cloudhut/common/logging"
	"github.com/cloudhut/owl-shop/pkg/shop"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"os"
	"strings"
)

type config struct {
	ConfigFilepath string
	Logger         logging.Config `yaml:"logger"`

	Shop shop.Config `yaml:"shop"`
}

func (c *config) SetDefaults() {
	c.Logger.SetDefaults()
	c.Shop.SetDefaults()
}

func (c *config) Validate() error {
	err := c.Logger.Set(c.Logger.LogLevelInput) // Parses LogLevel
	if err != nil {
		return fmt.Errorf("failed to validate loglevel input: %w", err)
	}

	return nil
}

// RegisterFlags for all (sub)configs
func (c *config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.ConfigFilepath, "config.filepath", "", "Path to the config file")
	c.Shop.RegisterFlags(f)
}

func LoadConfig(logger *zap.Logger) (config, error) {
	k := koanf.New(".")
	var cfg config
	cfg.SetDefaults()

	// Flags have to be parsed first because the yaml config filepath is supposed to be passed via flags
	flagext.RegisterFlags(&cfg)
	flag.Parse()

	// 1. Check if a config filepath is set via flags. If there is one we'll try to load the file using a YAML Parser
	var configFilepath string
	if cfg.ConfigFilepath != "" {
		configFilepath = cfg.ConfigFilepath
	} else {
		envKey := "CONFIG_FILEPATH"
		configFilepath = os.Getenv(envKey)
	}
	if configFilepath == "" {
		logger.Info("config filepath is not set, proceeding with options set from env variables and flags")
	} else {
		err := k.Load(file.Provider(configFilepath), yaml.Parser())
		if err != nil {
			return config{}, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	}

	// 2. Unmarshal the config into our config struct using the YAML and then ENV parser
	// We could unmarshal the loaded koanf input after loading both providers, however we want to unmarshal the YAML
	// config with `ErrorUnused` set to true, but unmarshal environment variables with `ErrorUnused` set to false (default).
	// Rationale: Orchestrators like Kubernetes inject unrelated environment variables, which we still want to allow.
	unmarshalCfg := koanf.UnmarshalConf{
		Tag:       "yaml",
		FlatPaths: false,
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc()),
			Metadata:         nil,
			Result:           &cfg,
			WeaklyTypedInput: true,
			ErrorUnused:      true,
			TagName:          "yaml",
		},
	}
	err := k.UnmarshalWithConf("", &cfg, unmarshalCfg)
	if err != nil {
		return config{}, fmt.Errorf("failed to unmarshal YAML config into config struct: %w", err)
	}

	err = k.Load(env.ProviderWithValue("", ".", func(s string, v string) (string, interface{}) {
		// key := strings.Replace(strings.ToLower(s), "_", ".", -1)
		key := strings.Replace(strings.ToLower(s), "_", ".", -1)
		// Check to exist if we have a configuration option already and see if it's a slice
		// If there is a comma in the value, split the value into a slice by the comma.
		if strings.Contains(v, ",") {
			return key, strings.Split(v, ",")
		}

		// Otherwise return the new key with the unaltered value
		return key, v
	}), nil)
	if err != nil {
		return config{}, fmt.Errorf("failed to unmarshal environment variables into config struct: %w", err)
	}

	unmarshalCfg.DecoderConfig.ErrorUnused = false
	err = k.UnmarshalWithConf("", &cfg, unmarshalCfg)
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}
