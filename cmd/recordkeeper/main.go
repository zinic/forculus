package main

import (
	"context"
	"flag"

	"github.com/zinic/forculus/config"

	"github.com/zinic/forculus/cmd"
	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/recordkeeper/server"
)

func start(cfg config.RecordKeeperConfig) error {
	if apiHandler, err := server.NewHandler(cfg); err != nil {
		log.Fatalf("Fatal error starting record keeper: %v", err)
	} else {
		serverInstance := server.NewServer(cfg, apiHandler)

		go func() {
			if err := serverInstance.ListenAndServe(); err != nil {
				log.Errorf("Fatal error while running HTTP server: %v", err)
			}
		}()

		cmd.WaitForSignal()

		log.Info("Shutting down")

		if err := serverInstance.Shutdown(context.Background()); err != nil {
			log.Errorf("Error during HTTP server shutdown: %v", err)
		}

		if err := apiHandler.Close(); err != nil {
			log.Errorf("Error during record keeper shutdown: %v", err)
		}

		log.Infof("Shutdown complete")
	}

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
	if cfg, err := config.LoadRecordKeeperCfg(cfgPath); err != nil {
		log.Fatalf("configuration error: %v", err)
	} else if validateCfg {
		log.Fatalf("Not implemented")
	} else if err := start(cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
