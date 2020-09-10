package config

import (
	"github.com/BurntSushi/toml"
)

func LoadConfiguration(path string) (Configuration, error) {
	cfg := configuration{}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Configuration{}, err
	}

	return parseConfigurationFields(cfg)
}
