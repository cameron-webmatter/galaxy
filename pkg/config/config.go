package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func LoadFromDir(dir string) (*Config, error) {
	configPath := filepath.Join(dir, "galaxy.config.toml")
	return Load(configPath)
}

func (c *Config) Validate() error {
	switch c.Output.Type {
	case OutputStatic, OutputServer, OutputHybrid:
	case "":
		c.Output.Type = OutputStatic
	default:
		return fmt.Errorf("invalid output type: %s (must be static, server, or hybrid)", c.Output.Type)
	}

	if c.Output.Type == OutputServer || c.Output.Type == OutputHybrid {
		if c.Adapter.Name == "" {
			c.Adapter.Name = AdapterStandalone
		}

		switch c.Adapter.Name {
		case AdapterStandalone, AdapterCloudflare, AdapterNetlify, AdapterVercel:
		default:
			return fmt.Errorf("invalid adapter: %s", c.Adapter.Name)
		}
	}

	if c.Server.Port == 0 {
		c.Server.Port = 4322
	}

	if c.Server.Host == "" {
		c.Server.Host = "localhost"
	}

	if c.OutDir == "" {
		c.OutDir = "./dist"
	}

	if c.Base == "" {
		c.Base = "/"
	}

	return nil
}

func (c *Config) IsSSR() bool {
	return c.Output.Type == OutputServer || c.Output.Type == OutputHybrid
}

func (c *Config) IsStatic() bool {
	return c.Output.Type == OutputStatic
}

func (c *Config) IsHybrid() bool {
	return c.Output.Type == OutputHybrid
}
