package main

import (
	"flag"

	"github.com/zinic/forculus/cmd"
	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/eventserver/actor"
	"github.com/zinic/forculus/eventserver/actor/logic"
	"github.com/zinic/forculus/eventserver/event"
	"github.com/zinic/forculus/eventserver/service"
	"github.com/zinic/forculus/log"
)

func start(cfg config.EventServerConfig) error {
	var (
		zmClient       = cmd.NewZoneminderClient(cfg.Zoneminder)
		serviceManager = service.NewManager()
		reactor        = actor.NewReactor(serviceManager)
		monitorWatch   = logic.NewMonitorWatch(zmClient, reactor)
	)

	if log.Thresholds().Accepts(log.LevelDebug) {
		logic.RegisterEventLogger(reactor)
	}

	logic.RegisterMonitorEventWatch(reactor, zmClient)

	for alertName, alertCfg := range cfg.EmailAlerts {
		logic.RegisterEventEmailSender(reactor, alertName, alertCfg, cfg.SMTPServers[alertCfg.Server])
	}

	for emailerName, emailerCfg := range cfg.Emailers {
		reactor.Register(logic.NewEmailSender(emailerName, emailerCfg, cfg.SMTPServers[emailerCfg.Server]))
	}

	if storageProviders, err := cmd.InitializeStorageProviders(cfg.StorageProviders); err != nil {
		log.Fatalf("Failed to initialize storage providers: %v", err)
	} else {
		for uploaderName, uploaderCfg := range cfg.Uploaders {
			var (
				provider = storageProviders[uploaderCfg.StorageTarget]
				uploader = logic.NewUploader(uploaderName, reactor, zmClient, provider, uploaderCfg)
			)

			reactor.Register(uploader, event.NewMonitorEvent)
			log.Debugf("New uploader %s registered to upload to storage provider %s", uploaderName, uploaderCfg.StorageTarget)
		}

	}

	serviceManager.Start(monitorWatch)

	cmd.WaitForSignal()

	log.Info("Shutting down")
	serviceManager.Stop()

	log.Info("Shutdown complete")
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

	if cfgPath == "" {
		log.Fatalf("Configuration path required.")
	}

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
	if cfg, err := config.LoadEventServerCfg(cfgPath); err != nil {
		log.Fatalf("configuration error: %v", err)
	} else if validateCfg {
		log.Fatalf("Not implemented")
	} else if err := start(cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
