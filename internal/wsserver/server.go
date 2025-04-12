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
	templateDir  = "web/templates/html"
	pingInterval = 25 * time.Second // интервал отправки Ping
	pongWait     = 30 * time.Second // время ожидания Pong
	writeWait    = 5 * time.Second  // таймаут записи
)

type WSServer interface {
	Start() error
	Stop() error
}

// для меня, тк не переваривалось
// в структуре мы просто определяем какие поля будут у типа
// далее говорим что будет в этом типе
// изначально там по сути ничего (выделяется ли под это память или нет?)
type wsSrv struct {
	mux        *http.ServeMux                          // Мультиплексор для маршрутизации HTTP-запросов
	srv        *http.Server                            // HTTP-сервер, который будет обрабатывать подключения
	wsUpg      websocket.Upgrader                      // Конфигурация для апгрейда HTTP -> WebSocket
	wsClients  map[*websocket.Conn]struct{}            // Мапа активных WebSocket-клиентов
	mutex      sync.RWMutex                            // т.к мапа ws.wsClients[conn] = struct{}{} является потоконебезопасной и возникнет гонка
	broadcast  chan *wsMsg                             // Канал для широковещательной рассылки сообщений ПЕРЕРАБОТАТЬ ДЛЯ КОНКРЕТНЫХ КЛИЕНТОВ
	rooms      map[string]map[*websocket.Conn]struct{} // комнаты: {"room1": {conn1, conn2}}
	roomsMutex sync.RWMutex
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
			HandshakeTimeout:  5 * time.Second, // время рукопожатия
			ReadBufferSize:    1024,            // буфер для чтения
			WriteBufferSize:   1024,            // буфер для записи
			WriteBufferPool:   nil,
			Subprotocols:      nil,
			Error:             nil,
			CheckOrigin:       nil, // !!! проверка домена (если nil, то разрешены все домены) !!!
			EnableCompression: false,
		},
		wsClients:  make(map[*websocket.Conn]struct{}),
		mutex:      sync.RWMutex{},
		broadcast:  make(chan *wsMsg),
		roomsMutex: sync.RWMutex{},
		rooms:      make(chan map[string]map[*websocket.Conn]struct{}),
	}
}

func (ws *wsSrv) Start() error {
	hub := newHub()
	go hub.run()
	ws.mux.Handle("/", http.FileServer(http.Dir(templateDir)))
	ws.mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.wsHandler(hub, w, r)
	})
	//go ws.writeToClientsBroadcast()
	//go ws.writeToClientBroadcast()
	return ws.srv.ListenAndServe()
}

// Stop метод структуры wsSrv интерфеса WSServer
// Выключает сервер (надо будет разобраться как отключать одно соединение, а не весь сервер
// Закрывает канал close(ws.broadcast)
// отправляет в цикле закрытия клиентов сообщение о закрытии (websocket.CloseMessage)
func (ws *wsSrv) Stop() error { // нужно подробно разобрать
	close(ws.broadcast)
	ws.mutex.Lock()
	// В цикле проходим по всем клиентам, выбираем на каждом шаге конкретное
	// соединение Далее работаем с конкретным соединением Для StudBridge надо будет
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
	}
	ws.mutex.Unlock()
	log.Info(ws.wsClients)
	return ws.srv.Shutdown(context.Background())
}

func (ws *wsSrv) wsHandler(hub, w http.ResponseWriter, r *http.Request) {
	conn, err := ws.wsUpg.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Error upgrading to websocket: %v", err)
		return
	}
	log.Infof("Client with address %s connected", conn.RemoteAddr().String())

	client := &Client{}

	go ws.readFromClient(conn)
	go ws.writeToClientBroadcast(conn)
}

func (ws *wsSrv) readFromClient(conn *websocket.Conn) {
	// удалили текущее соединение (клиента) в конце выаполенния функции
	defer func() {
		ws.mutex.Lock()
		delete(ws.wsClients, conn)
		ws.mutex.Unlock()
	}()

	for {
		msg := new(wsMsg)
		err := conn.ReadJSON(msg)
		if err != nil {
			log.Errorf("Error reading from websocket: %v", err)
			break
		}
		msg.IPAddress = conn.RemoteAddr().String()
		msg.Time = time.Now().Format("2006-01-02 15:04:05")
		ws.broadcast <- msg
	}

}

// writeToClientBroadcast - отвечает за управление исходящего трафика
// пинг сообщения, отправка сообщений клиенту
// такое ощущение, что утекает память...
func (ws *wsSrv) writeToClientBroadcast(conn *websocket.Conn) {
	// ticker - тикер, который срабатывает каждые pingInterval
	ticker := time.NewTicker(pingInterval)
	// Освобождение ресурсов после выхода
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case msg, ok := <-ws.broadcast:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteJSON(msg); err != nil {
				log.Infof("Error writing to client: %v", err)
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Infof("Error writing to client: %v", err)
				return
			}
		}
	}
}

// writeToClientsBroadcast - рассылка сообщений всем подключенным клиентам, а не конкретным
// идет под удаление
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
