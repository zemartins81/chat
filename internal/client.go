// internal/client.go

package internal

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

// Client é uma ponte entre a conexão WebSocket e o hub.
type Client struct {
	hub *Hub

	// A conexão WebSocket.
	conn *websocket.Conn

	// Canal de saída de mensagens.
	send chan []byte

	// ID e nome do usuário associado a este cliente.
	UserID   int
	Username string
	Room     *Room
}

// NewClient cria uma nova instância de Client.
// Esta função ajuda a manter a criação do cliente organizada.
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256), // Buffer para evitar bloqueios.
	}
}

// Register envia o cliente para o canal de registro do Hub.
func (c *Client) Register() {
	c.hub.register <- c
}

// ReadPump lê mensagens do WebSocket e as envia para o hub.
// Ele roda em sua própria goroutine para cada cliente.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var event Event
		if err := json.Unmarshal(rawMessage, &event); err != nil {
			log.Println("Erro ao desempacotar evento:", err)
			continue
		}

		switch event.Type {
		case "send_message":
			var payload SendMessagePayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				log.Println("Erro no payload de send_message:", err)
				continue
			}
			c.hub.broadcast <- &Message{
				RoomName: payload.Room,
				Content:  payload.Content,
				Username: c.Username,
			}

		// LÓGICA DO JOIN ROOM IMPLEMENTADA
		case "join_room":
			var payload JoinRoomPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				log.Println("Erro no payload de join_room:", err)
				continue
			}

			// Cria a requisição e envia para o canal 'join' do Hub.
			req := &JoinRequest{
				Client:   c,
				RoomName: payload.Room,
			}
			c.hub.join <- req

		default:
			log.Printf("Tipo de evento desconhecido: %s", event.Type)
		}
	}
}

// WritePump escreve mensagens do hub para a conexão WebSocket.
// Ele também roda em sua própria goroutine.
func (c *Client) WritePump() {
	// Garante que a conexão seja fechada ao sair.
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		// Aguarda por uma mensagem no canal 'send' do cliente.
		case message, ok := <-c.send:
			if !ok {
				// O Hub fechou o canal, então encerra a conexão.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Escreve a mensagem na conexão WebSocket.
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Adiciona mensagens restantes na fila ao mesmo writer para otimização.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'}) // Adiciona uma nova linha entre mensagens.
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
