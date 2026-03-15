package config

import _ "embed"

//go:embed defaults.yaml
var DefaultConfig string

type NINAConfig struct {
	Host string `yaml:"host"`
}

type AdapterConfig struct {
	Type string `yaml:"type"`
}

type AuthConfig struct {
	Token string `yaml:"token"`
}

type Config struct {
	ID          string        `yaml:"id"`
	DisplayName string        `yaml:"display_name"`
	NINA        NINAConfig    `yaml:"nina,omitempty"`
	Adapter     AdapterConfig `yaml:"adapter"`
	Auth        AuthConfig    `yaml:"auth,omitempty"`
}
