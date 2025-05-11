package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sbChat/internal/config"
	"sbChat/internal/database"
	"sbChat/internal/repository"
	"sbChat/internal/wsserver"
	"syscall"
	"time"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация БД
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализация схемы БД
	if err := db.InitSchema(); err != nil {
		log.Printf("Warning: schema initialization: %v", err)
	}

	// Создание репозитория
	chatRepo := repository.NewChatRepository(db.DB)

	// Создание WebSocket сервера
	wsServer := wsserver.NewWsServer(":8080", chatRepo)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting WebSocket server on :8080")
		if err := wsServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("WebSocket server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := wsServer.Stop(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	log.Println("Server stopped gracefully")
}
