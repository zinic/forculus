package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"path/filepath"
)

func LoadConfiguration(path string) (EventServerConfig, error) {
	cfg := configuration{}

	if absPath, err := filepath.Abs(path); err != nil {
		return EventServerConfig{}, fmt.Errorf("failed to resolve %s to an absolute path: %v\n", path, err)
	} else if _, err := toml.DecodeFile(absPath, &cfg); err != nil {
		return EventServerConfig{}, err
	}

	return parseConfigurationFields(cfg)
}
