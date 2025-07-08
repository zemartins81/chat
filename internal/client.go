package internal

import (
	"log"
	"github.com/gorilla/websocket"
)

type Client struct {
	hub *Hub

	conn *websocket.Conn

	send chan []byte
}

// NewClient cria um novo cliente
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
}

// ReadPump inicia o pump de leitura (método público)
func (c *Client) ReadPump() {
	c.readPump()
}

// WritePump inicia o pump de escrita (método público)
func (c *Client) WritePump() {
	c.writePump()
}

// Register registra o cliente no hub
func (c *Client) Register() {
	c.hub.register <- c
}

// readPump bombeia mensagens do WebSocket para o hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// Envia a mensagem lida para o canal de broadcast do Hub.
		c.hub.broadcast <- message
	}
}

// writePump bombeia mensagens do hub para a conexão WebSocket.
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// O hub fechou o canal.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
