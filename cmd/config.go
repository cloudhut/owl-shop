package main

import (
	"flag"
	"fmt"
	"github.com/cloudhut/common/logging"
	"github.com/cloudhut/owl-shop/pkg/shop"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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

// parseYAMLConfig reads the YAML-formatted config from filepath into cfg.
func parseYAMLConfig(filepath string, cfg *config) error {
	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	err = yaml.UnmarshalStrict(buf, cfg)
	if err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	return nil
}

func parseConfig() (*config, error) {
	cfg := &config{}
	cfg.SetDefaults()
	cfg.RegisterFlags(flag.CommandLine)
	flag.Parse()

	// This can not be part of the Config's validate() method because we haven't decoded the YAML config at this point
	if cfg.ConfigFilepath == "" {
		return nil, fmt.Errorf("you must specify the path to the config filepath using the --config.filepath flag")
	}

	err := parseYAMLConfig(cfg.ConfigFilepath, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load yaml config: %w", err)
	}
	err = cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}
