package websocket

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/generator"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/oapi"
	ws "github.com/gorilla/websocket"
)

type Websocket struct {
	connections map[string]*Connection
	Mutex       sync.RWMutex
	OnConnect   func(r *http.Request, conn *Connection) error
}

// New creates new Websocket instance
func New() *Websocket {
	connections := make(map[string]*Connection)
	return &Websocket{
		connections: connections,
		Mutex:       sync.RWMutex{},
	}
}

func (ws *Websocket) GetConnection(key string) (*Connection, bool) {
	c, ok := ws.connections[key]
	return c, ok
}

func (ws *Websocket) SendToAll(msgType, msg string) {
	ws.Mutex.RLock()
	conns := make([]*Connection, 0, len(ws.connections))
	for _, conn := range ws.connections {
		conns = append(conns, conn)
	}
	ws.Mutex.RUnlock()

	for _, conn := range conns {
		conn.Send(msgType, msg)
	}
}

func (ws *Websocket) CloseConnection(key string) {
	ws.Mutex.Lock()
	conn, ok := ws.connections[key]
	if ok {
		delete(ws.connections, key)
	}
	ws.Mutex.Unlock()
	if !ok {
		return
	}
	if conn.OnClose != nil {
		conn.OnClose()
	}
	select {
	case conn.closeChan <- true:
	default:
	}
	_ = conn.conn.Close()
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ws *Websocket) Handler(w http.ResponseWriter, r *http.Request) {
	wsb, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		log.Println("upgrade err:", err)
		oapi.ClientError(w, http.StatusBadRequest)
		return
	}

	// Create new connection
	// Assign new key
	k := generator.RandomSimpleString(18)
	conn := newConnection(k, wsb, ws)
	ws.Mutex.Lock()
	ws.connections[k] = conn
	ws.Mutex.Unlock()
	conn.startWriter()
	conn.conn.SetCloseHandler(func(code int, text string) error {
		ws.CloseConnection(k)
		return nil
	})

	if ws.OnConnect != nil {
		if err := ws.OnConnect(r, conn); err != nil {
			oapi.ServerError(w, err)
			return
		}
	}

	go func() {
		for {
			ws.Mutex.RLock()
			conn, ok := ws.connections[k]
			ws.Mutex.RUnlock()
			if !ok {
				break
			}

			_, r, err := conn.conn.NextReader()
			if err != nil {
				log.Println("websocket: reader:", err)
				ws.CloseConnection(k)
				break
			}

			bytes, err := io.ReadAll(r)
			if err != nil {
				log.Println("websocket:", err)
				ws.CloseConnection(k)
				break
			}

			var msg Message
			if err := json.Unmarshal(bytes, &msg); err == nil {
				if msg.Type == "DISCONNECT" {
					ws.CloseConnection(k)
					continue
				}
				if msg.Type == "PONG" {
					conn.isPonged = true
					continue
				}
				if conn.OnMessage != nil {
					conn.OnMessage(msg)
				}
			} else {
				if conn.OnBytes != nil {
					conn.OnBytes(bytes)
				}
			}
		}
	}()
}
