package ws

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Client struct {
	Hub  *WebSocketHub
	Conn *websocket.Conn
	Send chan []byte
	ID   string
}

func NewClient(hub *WebSocketHub, conn *websocket.Conn, id string) *Client {
	return &Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256),
		ID:   id,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			err := c.Conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}

}
