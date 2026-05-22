package websocket

import (
	"context"
	"time"

	"log"

	ws "github.com/gorilla/websocket"
)

type Message struct {
	Text string
	Type string
}

type Connection struct {
	Key            string
	conn           *ws.Conn
	messageQueue   chan Message
	OnMessage      func(Message)
	OnBytes        func([]byte)
	OnClose        func()
	closeChan      chan bool
	Context        context.Context
	isPonged       bool
	connectionPool *Websocket
}

func newConnection(key string, wsc *ws.Conn, websocket *Websocket) *Connection {
	return &Connection{
		conn:           wsc,
		messageQueue:   make(chan Message, 16),
		closeChan:      make(chan bool, 1),
		Key:            key,
		Context:        context.Background(),
		isPonged:       true,
		connectionPool: websocket,
	}
}

func (conn *Connection) Send(msgType, msg string) {
	select {
	case conn.messageQueue <- Message{Text: msg, Type: msgType}:
	case <-time.After(2 * time.Second):
		log.Println("websocket: message send timeout")
	}
}

// startWriter starts message writer from messageQueue AND starts pinger.
func (c *Connection) startWriter() {
	finish := false
	go func() {
		// TODO: Check performance on this. This loop might exhaust CPU.
		for {
			select {
			case msg := <-c.messageQueue:
				err := c.conn.WriteJSON(msg)
				if err != nil {
					log.Println("websocket: writer:", err)
					c.connectionPool.CloseConnection(c.Key)
					return
				}
			case <-c.closeChan:
				finish = true
			}
			if finish {
				break
			}
		}
	}()
	go func() {
		for {
			if !c.isPonged {
				c.connectionPool.CloseConnection(c.Key)
				return
			}
			select {
			case c.messageQueue <- Message{Type: "PING", Text: "Nothing"}:
				// c.isPonged = false // Uncomment it on prod
			case <-c.closeChan:
				finish = true
			}
			if finish {
				break
			}

			time.Sleep(5 * time.Second)
		}
	}()
}
