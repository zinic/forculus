package config

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

func LoadEventServerCfg(path string) (EventServerConfig, error) {
	cfg := eventServerConfiguration{}

	if absPath, err := filepath.Abs(path); err != nil {
		return EventServerConfig{}, fmt.Errorf("failed to resolve %s to an absolute path: %v\n", path, err)
	} else if _, err := toml.DecodeFile(absPath, &cfg); err != nil {
		return EventServerConfig{}, err
	}

	return parseEventServerCfg(cfg)
}

func LoadRecordKeeperCfg(path string) (RecordKeeperConfig, error) {
	cfg := RecordKeeperConfig{}

	if absPath, err := filepath.Abs(path); err != nil {
		return cfg, fmt.Errorf("failed to resolve %s to an absolute path: %v\n", path, err)
	} else if _, err := toml.DecodeFile(absPath, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
