package wsserver

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

const (
	templateDir = "web/templates/html"
)

type WSServer interface {
	Start() error
}

type wsSrv struct {
	mux       *http.ServeMux // Мультиплексор для маршрутизации HTTP-запросов
	srv       *http.Server   // HTTP-сервер, который будет обрабатывать подключения
	wsUpg     websocket.Upgrader
	wsClients map[*websocket.Conn]struct{}
	mutex     sync.RWMutex // т.к мапа ws.wsClients[conn] = struct{}{} является потоконебезопасной и возникнет гонка
	broadcast chan *wsMsg
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
			CheckOrigin:       nil,
			EnableCompression: false,
		},
		wsClients: make(map[*websocket.Conn]struct{}),
		mutex:     sync.RWMutex{},
		broadcast: make(chan *wsMsg),
	}
}

func (ws *wsSrv) Start() error {
	ws.mux.Handle("/", http.FileServer(http.Dir(templateDir)))
	ws.mux.HandleFunc("/ws", ws.wsHandler)
	ws.mux.HandleFunc("/test", ws.testHandler)
	go ws.writeToClientsBroadcast()
	return ws.srv.ListenAndServe()
}

func (ws *wsSrv) testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Test is successful"))
	log.Printf("Test is successful from %s", r.RemoteAddr)
}

func (ws *wsSrv) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.wsUpg.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Error upgrading to websocket: %v", err)
		return
	}
	log.Infof("Client with adress %s connected", conn.RemoteAddr().String())
	ws.mutex.Lock()
	ws.wsClients[conn] = struct{}{}
	ws.mutex.Unlock()
	ws.wsClients[conn] = struct{}{}
	go ws.readFromClient(conn)
}

func (ws *wsSrv) readFromClient(conn *websocket.Conn) {
	for {
		msg := new(wsMsg)
		err := conn.ReadJSON(msg)
		if err != nil {
			log.Errorf("Error reading from websocket: %v", err)
			break
		}
		//host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		//if err != nil {
		//	log.Errorf("Error reading from address split: %v", err)
		//}
		msg.IPAdress = conn.RemoteAddr().String()
		msg.Time = time.Now().Format("2006-01-02 15:04:05")
		ws.broadcast <- msg
	}
	// удалили текущее соединение (клиента)
	ws.mutex.Lock()
	delete(ws.wsClients, conn)
	ws.mutex.Unlock()
}

func (ws *wsSrv) writeToClientsBroadcast() {
	for msg := range ws.broadcast {
		ws.mutex.RLock()
		for client := range ws.wsClients {
			func() {
				if err := client.WriteJSON(msg); err != nil {
					log.Errorf("Error writing to client: %v", err)
				}
			}()
		}
		ws.mutex.RUnlock()
	}
}
