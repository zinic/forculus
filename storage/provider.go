package storage

import (
	"fmt"
	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/storage/aws"
	"io"
)

type Provider interface {
	Configure(cfg config.StorageProvider) error
	Validate(cfg config.StorageProvider) error
	Write(key string, reader io.Reader) error
}

func newProvider(provider config.StorageProviderType) (Provider, error) {
	switch provider {
	case config.ProviderAWS:
		return &aws.S3Provider{}, nil

	default:
		return nil, fmt.Errorf("unsupported provider type %s", provider)
	}
}

func New(cfg config.StorageProvider) (Provider, error) {
	if provider, err := newProvider(cfg.Provider); err != nil {
		return nil, err
	} else if err := provider.Configure(cfg); err != nil {
		return nil, err
	} else {
		return provider, nil
	}
}

func ValidateConfig(cfg config.StorageProvider) error {
	if provider, err := newProvider(cfg.Provider); err != nil {
		return err
	} else {
		return provider.Validate(cfg)
	}
}
