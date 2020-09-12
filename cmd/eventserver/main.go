package main

import (
	"flag"
	"fmt"
	"github.com/zinic/forculus/eventserver/event"
	"github.com/zinic/forculus/storage"
	"sync"

	"github.com/zinic/forculus/cmd"
	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/eventserver/actor"
	"github.com/zinic/forculus/eventserver/actor/logic"
	"github.com/zinic/forculus/log"
)

func start(cfg config.EventServerConfig) error {
	var (
		zmClient     = cmd.NewZoneminderClient(cfg.Zoneminder)
		waitGroup    = &sync.WaitGroup{}
		reactor      = actor.NewReactor()
		monitorWatch = logic.NewMonitorWatch(zmClient, reactor)
	)

	if log.Thresholds().Accepts(log.LevelDebug) {
		logic.RegisterEventLogger(reactor)
	}

	logic.RegisterMonitorEventWatch(reactor, zmClient)

	for alertName, alertCfg := range cfg.EmailAlerts {
		logic.RegisterEmailer(reactor, alertName, alertCfg, cfg.SMTPServers[alertCfg.Server])
	}

	storageProviders := make(map[string]storage.Provider)
	for providerName, storageProviderCfg := range cfg.StorageProviders {
		if provider, err := storage.New(storageProviderCfg); err != nil {
			return fmt.Errorf("failed initializing storage provider %s: %w", providerName, err)
		} else {
			storageProviders[providerName] = provider
		}

		log.Debugf("New storage provider %s type %s registered", providerName, storageProviderCfg.Provider)
	}

	for uploaderName, uploaderCfg := range cfg.Uploaders {
		var (
			provider = storageProviders[uploaderCfg.StorageTarget]
			uploader = logic.NewUploader(uploaderName, zmClient, provider, uploaderCfg)
		)

		reactor.Register(uploader, event.NewMonitorEvent)
		log.Debugf("New uploader %s registered to upload to storage provider %s", uploaderName, uploaderCfg.StorageTarget)
	}

	reactor.Start(waitGroup)
	monitorWatch.Start(waitGroup)

	cmd.WaitForSignal()

	log.Info("Shutting down")

	reactor.Stop()
	monitorWatch.Stop()

	waitGroup.Wait()

	log.Info("Shut down complete")
	return nil
}

func main() {
	var (
		cfgPath     string
		validateCfg bool
		enableInfo  bool
		enableDebug bool
	)

	flag.StringVar(&cfgPath, "c", "", "Path to a valid configuration TOML file.")
	flag.BoolVar(&enableInfo, "v", false, "Enable verbose output.")
	flag.BoolVar(&enableDebug, "d", false, "Enable debug output. This switch supersedes verbose output.")
	flag.BoolVar(&validateCfg, "validate", false, "Validate configuration.")
	flag.Parse()

	// Start with log configuration defaults
	log.ConfigureDefaults()

	// Configure log output
	log.Configure()

	if enableInfo {
		log.AddOutput(log.NewStdoutLogger(log.LevelInfo, ""))
	} else if enableDebug {
		log.AddOutput(log.NewStdoutLogger(log.LevelDebug, ""))
	} else {
		log.AddOutput(log.NewStdoutLogger(log.LevelWarn, ""))
	}

	// Load configuration and either output that it's valid or start the daemon
	if cfg, err := config.LoadConfiguration(cfgPath); err != nil {
		log.Fatalf("configuration error: %v", err)
	} else if validateCfg {
		log.Fatalf("Not implemented")
	} else if err := start(cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
