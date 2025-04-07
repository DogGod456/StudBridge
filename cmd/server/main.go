package main

import (
	"StudBridge/internal/wsserver"
	log "github.com/sirupsen/logrus"
)

const (
	addr = "192.168.0.150:9090"
)

func main() {
	wsSrv := wsserver.NewWsServer(addr)
	log.Info("Starting server...")
	if err := wsSrv.Start(); err != nil {
		log.Errorf("Error starting wsServer: %v", err)
	}
}
