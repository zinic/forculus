package cmd

import (
	"fmt"
	"github.com/zinic/forculus/storage/providers"
	"os"
	"os/signal"
	"syscall"

	"github.com/zinic/forculus/apitools"

	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/storage"

	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/zoneminder/zmapi"
)

func InitializeStorageProviders(storageProviderCfgs map[string]config.StorageProvider) (map[string]storage.Provider, error) {
	storageProviders := make(map[string]storage.Provider)
	for providerName, storageProviderCfg := range storageProviderCfgs {
		if provider, err := providers.New(storageProviderCfg); err != nil {
			return nil, fmt.Errorf("failed initializing storage provider %s: %w", providerName, err)
		} else {
			storageProviders[providerName] = provider
		}

		log.Debugf("New storage provider %s type %s registered", providerName, storageProviderCfg.Provider)
	}

	return storageProviders, nil
}

func NewZoneminderClient(cfg config.Zoneminder) zmapi.Client {
	var (
		endpoint = apitools.NewEndpoint(
			cfg.Scheme,
			cfg.Host,
			cfg.Port,
			cfg.RootPath)

		credentials = zmapi.LoginCredentials{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	)

	return zmapi.NewClient(endpoint, credentials)
}

func WaitForSignal() {
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGINT, syscall.SIGTERM)

	<-signalC
}
