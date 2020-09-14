package main

import (
	"flag"
	"time"

	"github.com/zinic/forculus/cmd"
	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/log"
)

func main() {
	var (
		cfgPath string
	)

	flag.StringVar(&cfgPath, "c", "", "Path to a valid configuration TOML file.")
	flag.Parse()

	// Configure log defaults
	log.Configure()
	log.AddOutput(log.NewStdoutLogger(log.LevelDebug, ""))

	// If the configuration path is empty let the user know what's up
	if cfgPath == "" {
		log.Fatal("No configuration specified.")
	}

	// Load configuration and either output that it's valid or start the daemon
	if cfg, err := config.LoadEventServerCfg(cfgPath); err != nil {
		log.Fatalf("configuration error: %v", err)
	} else {
		client := cmd.NewZoneminderClient(cfg.Zoneminder)

		if err := client.Login(); err != nil {
			log.Fatalf("Error logging in: %v", err)
		}

		session := client.LoginSession()
		nextRefresh := session.LastRefresh.Add(time.Second * time.Duration(session.Details.AccessTokenExpires))
		nextLogin := session.Created.Add(time.Second * time.Duration(session.Details.RefreshTokenExpires))

		log.Infof("Login session")
		log.Infof("   Last refresh %s", session.LastRefresh)
		log.Infof("   Next refresh %s", nextRefresh)
		log.Infof("   Next login   %s", nextLogin)
	}
}
