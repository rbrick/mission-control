package adapter

import (
	"fmt"

	"github.com/rbrick/mission-control/rig/config"
)

func FromConfig(cfg *config.Config) ([]Adapter, error) {
	configs := cfg.Adapters
	if len(configs) == 0 {
		legacy := cfg.Adapter
		if legacy.ID == "" {
			legacy.ID = legacy.Type
		}
		if legacy.Host == "" {
			legacy.Host = cfg.NINA.Host
		}
		configs = []config.AdapterConfig{legacy}
	}
	if len(configs) == 0 || configs[0].Type == "" {
		configs = []config.AdapterConfig{{ID: "sim", Type: "sim"}}
	}

	adapters := make([]Adapter, 0, len(configs))
	for _, adapterConfig := range configs {
		switch adapterConfig.Type {
		case "", "sim", "simulated":
			adapters = append(adapters, NewSim(adapterConfig.ID))
		case "nina":
			adapter, err := NewNINA(adapterConfig.ID, adapterConfig.Host)
			if err != nil {
				return nil, err
			}
			adapters = append(adapters, adapter)
		default:
			return nil, fmt.Errorf("unsupported adapter type %q", adapterConfig.Type)
		}
	}
	return adapters, nil
}
