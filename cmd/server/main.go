package main

import (
	"StudBridge/internal/wsserver"
	log "github.com/sirupsen/logrus"
)

func main() {
	wsSrv := wsserver.NewWsServer(":8080")
	log.Info("Starting server...")
	if err := wsSrv.Start(); err != nil {
		log.Errorf("Error starting wsServer: %v", err)
	}
}
