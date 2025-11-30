package websocket

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"

	ws "github.com/gorilla/websocket"
	"github.com/peakdot/go-nuxt-example/backend/pkg/common/generator"
	"github.com/peakdot/go-nuxt-example/backend/pkg/common/oapi"
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
	for _, conn := range ws.connections {
		conn.Send(msgType, msg)
	}
}

func (ws *Websocket) CloseConnection(key string) {
	if _, ok := ws.connections[key]; !ok {
		return
	}
	if ws.connections[key].OnClose != nil {
		ws.connections[key].OnClose()
	}
	ws.connections[key].closeChan <- true
	ws.connections[key].conn.Close()
	ws.Mutex.Lock()
	delete(ws.connections, key)
	ws.Mutex.Unlock()
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
	ws.Mutex.Lock()
	ws.connections[k] = newConnection(k, wsb, ws)
	ws.Mutex.Unlock()
	ws.connections[k].startWriter()
	ws.connections[k].conn.SetCloseHandler(func(code int, text string) error {
		ws.CloseConnection(k)
		return nil
	})

	if ws.OnConnect != nil {
		if err := ws.OnConnect(r, ws.connections[k]); err != nil {
			oapi.ServerError(w, err)
			return
		}
	}

	go func() {
		for {
			if _, ok := ws.connections[k]; !ok {
				break
			}

			_, r, err := ws.connections[k].conn.NextReader()
			if err != nil {
				log.Println("websocket: reader:", err)
				break
			}

			bytes, err := io.ReadAll(r)
			if err != nil {
				log.Println("websocket:", err)
				break
			}

			var msg Message
			if err := json.Unmarshal(bytes, &msg); err == nil {
				if msg.Type == "DISCONNECT" {
					ws.CloseConnection(k)
					continue
				}
				if msg.Type == "PONG" {
					ws.connections[k].isPonged = true
					continue
				}
				if ws.connections[k].OnMessage != nil {
					ws.connections[k].OnMessage(msg)
				}
			} else {
				if ws.connections[k].OnBytes != nil {
					ws.connections[k].OnBytes(bytes)
				}
			}
		}
	}()
}
