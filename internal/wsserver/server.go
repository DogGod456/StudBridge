package wsserver

import (
	"context"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

const (
	templateDir    = "web/templates/html"
	pingInterval   = 5 * time.Second  // интервал отправки Ping
	pongWait       = 10 * time.Second // время ожидания Pong
	writeWait      = 5 * time.Second  // таймаут записи
	maxMessageSize = 1024 * 1024      // максимальный размер сообщения 1 Мб
)

type WSServer interface {
	Start() error
	Stop() error
}

type wsSrv struct {
	mux       *http.ServeMux // Мультиплексор для маршрутизации HTTP-запросов
	srv       *http.Server   // HTTP-сервер, который будет обрабатывать подключения
	wsUpg     websocket.Upgrader
	wsClients map[*websocket.Conn]struct{}
	mutex     sync.RWMutex // т.к мапа ws.wsClients[conn] = struct{}{} является потоконебезопасной и возникнет гонка
	broadcast chan *wsMsg
	done      chan struct{} // для коректного закрытия горутин, чтобы они не весели
}

func NewWsServer(addr string) WSServer {
	m := http.NewServeMux()
	return &wsSrv{
		mux: m,
		srv: &http.Server{
			Addr:    addr,
			Handler: m,
		},
		wsUpg: websocket.Upgrader{
			HandshakeTimeout:  5 * time.Second,
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			WriteBufferPool:   nil,
			Subprotocols:      nil,
			Error:             nil,
			CheckOrigin:       nil, // чекнуть что это такое и почему в проде надо смотреть на это
			EnableCompression: false,
		},
		wsClients: make(map[*websocket.Conn]struct{}),
		mutex:     sync.RWMutex{},
		broadcast: make(chan *wsMsg),
		done:      make(chan struct{}),
	}
}

func (ws *wsSrv) Start() error {
	ws.mux.HandleFunc("/ws", ws.wsHandler)
	go ws.writeToClientsBroadcast()
	log.Infof("Starting pure WebSocket server on %s", ws.srv.Addr)
	return ws.srv.ListenAndServe()
}

// Stop метод структуры wsSrv интерфеса WSServer
// Выключает сервер (надо будет разобраться как отключать одно соединение, а не весь сервер
// Закрывает канал close(ws.broadcast)
// отправляет в цикле закрытия клиентов сообщение о закрытии (websocket.CloseMessage)
func (ws *wsSrv) Stop() error { // нужно подробно разобрать
	close(ws.done)
	close(ws.broadcast)

	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	// В цикле проходим по всем клиентам, выбираем на каждом шаге конкретное
	// соединение Далее работаем с конкретным соединением Для sb надо будет
	// брать из таблицы два ip и закрывать их (1 комната)
	// (как я полагаю, возможно что то поменяется)
	for conn := range ws.wsClients {
		// Отправляем сообщение о корректном закрытии соединения
		conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(writeWait),
		)
		// Закрываем соединение
		conn.Close()
		// удаляем
		log.Infof("info about connection: %v", conn)
		delete(ws.wsClients, conn)
		log.Infof("Closed connection: %v", conn.RemoteAddr())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Info(ws.wsClients)
	return ws.srv.Shutdown(ctx)
}

func (ws *wsSrv) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.wsUpg.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Error upgrading to websocket: %v", err)
		return
	}

	log.Infof("Client with address %s connected", conn.RemoteAddr().String())

	ws.mutex.Lock()
	ws.wsClients[conn] = struct{}{}
	ws.mutex.Unlock()

	go ws.readFromClient(conn)
}

func (ws *wsSrv) readFromClient(conn *websocket.Conn) {
	// удалили текущее соединение (клиента) в конце выполнения функции
	defer func() {
		ws.mutex.Lock()
		delete(ws.wsClients, conn)
		ws.mutex.Unlock()
		conn.Close()
	}()

	for {
		msg := new(wsMsg)
		if err := conn.ReadJSON(msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Errorf("Read error: %v", err)
			}
			break
		}

		msg.IPAddress = conn.RemoteAddr().String()
		msg.Time = time.Now().Format(time.RFC3339)
		select {
		case ws.broadcast <- msg:
		case <-ws.done:
			return
		}
	}

}

func (ws *wsSrv) writeToClientsBroadcast() {
	for msg := range ws.broadcast {
		ws.mutex.RLock()
		for client := range ws.wsClients {
			go func() {
				if err := client.WriteJSON(msg); err != nil {
					log.Errorf("Error writing to client: %v", err)
				}
			}()
		}
		ws.mutex.RUnlock()
	}
}
