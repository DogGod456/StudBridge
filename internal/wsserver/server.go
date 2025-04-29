package wsserver

import (
	"context"
	"log"
	"net/http"
	"sbChat/internal/models"
	"sbChat/internal/repository"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WsServer интерфейс для работы с WebSocket сервером
// Содержит методы для запуска и остановки сервера
type WsServer interface {
	Start() error                   // Запускает сервер
	Stop(ctx context.Context) error // Останавливает сервер с учетом контекста
}

// wsServer реализация WebSocket сервера
type wsServer struct {
	server    *http.Server                 // HTTP сервер для обработки соединений
	upgrader  websocket.Upgrader           // Преобразователь HTTP в WebSocket
	clients   map[*websocket.Conn]struct{} // Мапа подключенных клиентов
	mu        sync.RWMutex                 // RWMutex для безопасного доступа к clients
	broadcast chan models.Message          // Канал для трансляции сообщений клиентам
	repo      repository.ChatRepository    // Репозиторий для работы с сообщениями в БД
}

// NewWsServer создает новый экземпляр WebSocket сервера
// addr - адрес сервера (например ":8080")
// repo - репозиторий для сохранения сообщений
func NewWsServer(addr string, repo repository.ChatRepository) WsServer {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	ws := &wsServer{
		server: srv,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024, // Размер буфера для чтения
			WriteBufferSize: 1024, // Размер буфера для записи
			CheckOrigin: func(r *http.Request) bool {
				return true // В продакшене нужно реализовать проверку origin!
			},
		},
		clients:   make(map[*websocket.Conn]struct{}), // Инициализация мапы клиентов
		broadcast: make(chan models.Message, 256),     // Буферизированный канал на 256 сообщений
		repo:      repo,                               // Сохраняем переданный репозиторий
	}

	// Регистрируем обработчик WebSocket соединений
	mux.HandleFunc("/ws", ws.handleConnection)
	return ws
}

// Start запускает WebSocket сервер
// Запускает в отдельной горутине обработчик broadcast сообщений
// Возвращает ошибку если сервер не смог запуститься
func (ws *wsServer) Start() error {
	go ws.handleBroadcasts() // Запускаем обработчик рассылки сообщений
	return ws.server.ListenAndServe()
}

// Stop корректно останавливает WebSocket сервер
// ctx - контекст для контроля времени остановки
// Закрывает все соединения и останавливает сервер
func (ws *wsServer) Stop(ctx context.Context) error {
	close(ws.broadcast) // Закрываем канал broadcast для остановки горутин

	ws.mu.Lock() // Блокируем запись для безопасной работы с clients
	defer ws.mu.Unlock()

	// Закрываем все активные соединения
	for client := range ws.clients {
		// Отправляем сообщение о закрытии соединения
		client.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		client.Close() // Закрываем соединение
	}

	return ws.server.Shutdown(ctx) // Останавливаем HTTP сервер
}

// handleConnection обрабатывает новое WebSocket соединение
// w - HTTP ResponseWriter
// r - HTTP Request
func (ws *wsServer) handleConnection(w http.ResponseWriter, r *http.Request) {
	// Обновляем HTTP соединение до WebSocket
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Добавляем новое соединение в мапу клиентов
	ws.mu.Lock()
	ws.clients[conn] = struct{}{}
	ws.mu.Unlock()

	// Гарантируем удаление соединения при выходе из функции
	defer func() {
		ws.mu.Lock()
		delete(ws.clients, conn) // Удаляем соединение из мапы
		ws.mu.Unlock()
		conn.Close() // Закрываем соединение
	}()

	// Читаем сообщения от клиента в цикле
	for {
		var msg models.Message
		// Читаем JSON сообщение от клиента
		if err := conn.ReadJSON(&msg); err != nil {
			// Обрабатываем только неожиданные ошибки закрытия
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("Read error: %v", err)
			}
			break
		}

		// Сохраняем сообщение в БД через репозиторий
		msgID, err := ws.repo.SendMessage(r.Context(), msg.ChatID, msg.SenderID, msg.Text, false)
		if err != nil {
			log.Printf("Failed to save message: %v", err)
			continue
		}

		// Обновляем сообщение данными из БД
		msg.ID = msgID
		msg.SendingTime = time.Now()

		// Отправляем сообщение в канал для трансляции
		ws.broadcast <- msg
	}
}

// handleBroadcasts рассылает сообщения всем подключенным клиентам
// Работает в отдельной горутине до закрытия канала broadcast
func (ws *wsServer) handleBroadcasts() {
	// Читаем сообщения из канала пока он не закрыт
	for msg := range ws.broadcast {
		ws.mu.RLock() // Блокируем на чтение

		// Рассылаем сообщение всем клиентам
		for client := range ws.clients {
			go func(c *websocket.Conn) {
				// Пытаемся отправить сообщение
				if err := c.WriteJSON(msg); err != nil {
					log.Printf("Write error: %v", err)
					// При ошибке удаляем клиента
					ws.mu.Lock()
					delete(ws.clients, c)
					ws.mu.Unlock()
					c.Close() // Закрываем проблемное соединение
				}
			}(client)
		}

		ws.mu.RUnlock() // Разблокируем чтение
	}
}
