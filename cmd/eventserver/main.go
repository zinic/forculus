package main

import (
	"flag"
	"github.com/zinic/forculus/eventserver"
	"github.com/zinic/forculus/eventserver/services"

	"github.com/zinic/forculus/cmd"
	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/eventserver/actors"
	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/service"
)

func start(cfg config.EventServerConfig) error {
	var (
		zmClient       = cmd.NewZoneminderClient(cfg.Zoneminder)
		serviceManager = service.NewManager()
		reactor        = eventserver.NewDispatch(serviceManager)
		monitorWatch   = services.NewMonitorWatch(zmClient, reactor)
	)

	if log.Thresholds().Accepts(log.LevelDebug) {
		actors.RegisterEventLogger(reactor)
	}

	actors.RegisterMonitorEventWatch(reactor, zmClient)

	for alertName, alertCfg := range cfg.EmailAlerts {
		actors.RegisterEventEmailSender(reactor, alertName, alertCfg, cfg.SMTPServers[alertCfg.Server])

		log.Debugf("New email alert %s registered to send to SMTP server %s", alertName, alertCfg.Server)
	}

	if storageProviders, err := cmd.InitializeStorageProviders(cfg.StorageProviders); err != nil {
		log.Fatalf("Failed to initialize storage providers: %v", err)
	} else {
		for uploaderName, uploaderCfg := range cfg.Uploaders {
			var (
				provider = storageProviders[uploaderCfg.StorageTarget]
				uploader = actors.NewUploader(uploaderName, reactor, zmClient, provider, uploaderCfg)
			)

			reactor.Register(uploader, eventserver.MonitorNewEvent)
			
			log.Debugf("New uploader %s registered to upload to storage provider %s", uploaderName, uploaderCfg.StorageTarget)
		}

	}

	for recordKeeperName, recordKeeperCfg := range cfg.RecordKeepers {
		reactor.Register(actors.NewRecordKeeper(reactor, recordKeeperCfg), eventserver.MonitorEventUploaded)

		log.Debugf("New record keeper %s registered", recordKeeperName)
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
