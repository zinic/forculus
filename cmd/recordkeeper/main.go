package main

import (
	"context"
	"github.com/zinic/forculus/cmd"
	"github.com/zinic/forculus/log"
	"github.com/zinic/forculus/recordkeeper/api"
)

func main() {
	log.Configure()
	log.AddOutput(log.NewStdoutLogger(log.LevelDebug, ""))

	if apiHandler, err := api.NewHandler(); err != nil {
		log.Fatalf("Fatal error starting record keeper: %v", err)
	} else {
		server := api.NewServer("0.0.0.0:8080", apiHandler)

		go func() {
			if err := server.ListenAndServe(); err != nil {
				log.Errorf("Fatal error while running HTTP server: %v", err)
			}
		}()

		cmd.WaitForSignal()

		if err := server.Shutdown(context.Background()); err != nil {
			log.Errorf("Error during HTTP server shutdown: %v", err)
		}

		if err := apiHandler.Close(); err != nil {
			log.Errorf("Error during record keeper shutdown: %v", err)
		}
	}
}
