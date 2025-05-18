package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sbChat/internal/config"
	"sbChat/internal/database"
	handlerChat "sbChat/internal/handler/chat"
	"sbChat/internal/repository"
	usecaseChat "sbChat/internal/usecase/chat"
	usecaseParticipant "sbChat/internal/usecase/participant"
	usecaseParticipantType "sbChat/internal/usecase/participant_type"
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
	participantRepo := repository.NewParticipantRepository(db.DB)
	participantTypeRepo := repository.NewParticipantTypeRepository(db.DB)

	// Инициализация use case для создания типа участника
	participantTypeUC := usecaseParticipantType.NewCreateParticipantTypeUseCase(participantTypeRepo)

	// Инициализация use case для работы с участниками
	participantUC := usecaseParticipant.NewParticipantUseCase(
		participantRepo,
		participantTypeUC,
	)

	// Инициализация use case для создания чата
	createChatUseCase := usecaseChat.NewCreateChatUseCase(
		chatRepo,
		participantUC,
	)

	// Создание HTTP хендлера для создания чатов
	createChatHandler := handlerChat.NewCreateChatHandler(createChatUseCase)

	// Настройка HTTP маршрутов
	mux := http.NewServeMux()
	mux.Handle("/api/chats/create", createChatHandler)

	// Создание WebSocket сервера
	wsServer := wsserver.NewWsServer(":8080", chatRepo)

	// Запуск HTTP сервера для API
	apiServer := &http.Server{
		Addr:    ":8081", // или другой порт, отличный от WebSocket
		Handler: mux,
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting WebSocket server on :8080")
		if err := wsServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("WebSocket server error: %v", err)
		}
	}()

	go func() {
		log.Println("Starting API server on :8081")
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Остановка WebSocket сервера
	if err := wsServer.Stop(ctx); err != nil {
		log.Printf("Error during WebSocket server shutdown: %v", err)
	}

	// Остановка HTTP сервера
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Error during API server shutdown: %v", err)
	}

	log.Println("Servers stopped gracefully")
}
