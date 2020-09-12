package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/zoneminder/api"
)

func NewZoneminderClient(cfg config.Zoneminder) api.Client {
	var (
		endpoint = api.NewEndpoint(
			cfg.Scheme,
			cfg.Host,
			cfg.Port,
			cfg.RootPath)

		credentials = api.LoginCredentials{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	)

	return api.NewClient(endpoint, credentials)
}

func WaitForSignal() {
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGINT, syscall.SIGTERM)

	<-signalC
}
