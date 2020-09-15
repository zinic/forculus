package providers

import (
	"fmt"

	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/storage"
	"github.com/zinic/forculus/storage/providers/aws"
)

func newProvider(provider config.StorageProviderType) (storage.Provider, error) {
	switch provider {
	case config.ProviderAWS:
		return &aws.S3Provider{}, nil

	default:
		return nil, fmt.Errorf("unsupported provider type %s", provider)
	}
}

func New(cfg config.StorageProvider) (storage.Provider, error) {
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
