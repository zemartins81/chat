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
	// Garante que o cliente seja desregistrado e a conexão fechada ao sair.
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		// Lê uma mensagem da conexão WebSocket.
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			// Se houver um erro (ex: cliente fechou a aba), sai do loop.
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// AQUI ESTÁ A LÓGICA NOVA:
		// 1. Cria a mensagem estruturada com o nome de usuário e o conteúdo.
		msg := Message{
			Username: c.Username,
			Content:  string(rawMessage),
		}

		// 2. Converte (marshal) a struct para JSON.
		jsonMessage, err := json.Marshal(msg)
		if err != nil {
			log.Println("Erro ao converter mensagem para JSON:", err)
			continue // Pula para a próxima iteração se houver erro.
		}

		// 3. Envia a mensagem JSON para o canal de broadcast do Hub.
		c.hub.broadcast <- jsonMessage
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
