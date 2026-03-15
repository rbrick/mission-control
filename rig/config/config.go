package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

//go:embed defaults.yaml
var DefaultConfig string

type NINAConfig struct {
	Host string `mapstructure:"host" yaml:"host"`
}

type AdapterConfig struct {
	Type string `mapstructure:"type" yaml:"type"`
}

type AuthConfig struct {
	Token string `mapstructure:"token" yaml:"token"`
}

type Config struct {
	ID          string        `mapstructure:"id" yaml:"id"`
	DisplayName string        `mapstructure:"display_name" yaml:"display_name"`
	NINA        NINAConfig    `mapstructure:"nina" yaml:"nina,omitempty"`
	Adapter     AdapterConfig `mapstructure:"adapter" yaml:"adapter"`
	Auth        AuthConfig    `mapstructure:"auth" yaml:"auth,omitempty"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	if err := v.ReadConfig(bytes.NewBufferString(DefaultConfig)); err != nil {
		return nil, fmt.Errorf("read embedded default config: %w", err)
	}

	if path != "" {
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config file %q: %w", path, err)
		}

		if err := v.MergeConfig(bytes.NewBuffer(contents)); err != nil {
			return nil, fmt.Errorf("merge config file %q: %w", path, err)
		}
	}

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("decode rig config: %w", err)
	}

	return config, nil
}
