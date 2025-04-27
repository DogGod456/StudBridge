package main

import (
	log "github.com/sirupsen/logrus"
	"sbChat/internal/wsserver"
)

const (
	addr = "192.168.0.150:9090"
)

func main() {
	wsSrv := wsserver.NewWsServer(":8080")
	log.Info("Starting server...")
	if err := wsSrv.Start(); err != nil {
		log.Errorf("Error starting wsServer: %v", err)
	}
}
