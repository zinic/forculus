package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/zinic/forculus/actor"
	"github.com/zinic/forculus/actor/logic"
	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/zoneminder/api"
)

func loadConfiguration(path string) (config.Configuration, error) {
	if absPath, err := filepath.Abs(path); err != nil {
		return config.Configuration{}, fmt.Errorf("failed to resolve %s to an absolute path: %v\n", path, err)
	} else if cfg, err := config.LoadConfiguration(absPath); err != nil {
		return config.Configuration{}, fmt.Errorf("parsing configuration %s failed: %v\n", path, err)
	} else {
		return cfg, nil
	}
}

func newZoneminderClient(cfg config.Zoneminder) api.Client {
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

func waitForSignal() {
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGINT, syscall.SIGTERM)

	<-signalC
}

func start(cfg config.Configuration) error {
	var (
		client       = newZoneminderClient(cfg.Zoneminder)
		waitGroup    = &sync.WaitGroup{}
		reactor      = actor.NewReactor()
		monitorWatch = logic.NewMonitorWatch(client, reactor)
	)

	if log.Thresholds().Accepts(log.LevelDebug) {
		logic.RegisterEventLogger(reactor)
	}

	logic.RegisterMonitorEventWatch(reactor, client)

	for _, emailAlert := range cfg.EmailAlerts {
		logic.RegisterEmailer(reactor, emailAlert, cfg.SMTPServer[emailAlert.Server])
	}

	for _, s3UploaderCfg := range cfg.S3Uploaders {
		logic.RegisterEventS3Uploader(reactor, client, s3UploaderCfg)
	}

	reactor.Start(waitGroup)
	monitorWatch.Start(waitGroup)

	waitForSignal()

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

	// If the configuration path is empty let the user know what's up
	if cfgPath == "" {
		log.Fatal("No configuration specified.")
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
	if cfg, err := loadConfiguration(cfgPath); err != nil {
		log.Fatalf("configuration error: %v", err)
	} else if validateCfg {
		outputCfgDetails(cfg)
	} else if err := start(cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func outputCfgDetails(cfg config.Configuration) {
	for s3UploaderName, s3Uploader := range cfg.S3Uploaders {
		log.Warnf("S3 Uploader: %s", s3UploaderName)
		log.Warnf("   Region:                %s", s3Uploader.Region)
		log.Warnf("   Bucket:                %s", s3Uploader.Bucket)
		log.Warnf("   Name Filter:           %s", s3Uploader.Filter.NameRegex)
		log.Warnf("   Alert Frame Threshold: %d", s3Uploader.Filter.AlertFrameThreshold)
		log.Warn("")
	}

	for emailAlertName, emailAlertCfg := range cfg.EmailAlerts {
		outputBuilder := strings.Builder{}
		outputBuilder.WriteString(fmt.Sprintf("Email Alert: %s\n", emailAlertName))
		outputBuilder.WriteString(fmt.Sprintf("   Name Filter:           %s\n", emailAlertCfg.Filter.NameRegex))
		outputBuilder.WriteString(fmt.Sprintf("   Event Trigger:         %s\n", emailAlertCfg.Filter.EventTrigger))
		outputBuilder.WriteString(fmt.Sprintf("   Alert Frame Threshold: %d\n", emailAlertCfg.Filter.AlertFrameThreshold))
		outputBuilder.WriteString("\n")

		log.Warn(outputBuilder.String())
	}
}
