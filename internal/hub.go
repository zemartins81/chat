package internal

import "log"

type Room struct {
	Name    string
	clients map[*Client]bool
}

// Nova struct para a requisição de entrar na sala
type JoinRequest struct {
	Client   *Client
	RoomName string
}

type Hub struct {
	rooms      map[string]*Room
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	join       chan *JoinRequest // NOVO CANAL
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
		join:       make(chan *JoinRequest), // Inicializa o novo canal
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			log.Printf("Cliente registrado: %s", client.Username)

		case client := <-h.unregister:
			// Se o cliente estava em uma sala, remove-o.
			if client.Room != nil {
				delete(client.Room.clients, client)
			}
			log.Printf("Cliente removido: %s", client.Username)

		// NOVO CASE para lidar com a entrada em salas
		case req := <-h.join:
			client := req.Client
			roomName := req.RoomName

			// 1. Remove o cliente da sala antiga, se ele estiver em uma.
			if client.Room != nil {
				delete(client.Room.clients, client)
			}

			// 2. Procura a sala de destino. Se não existir, cria.
			room, ok := h.rooms[roomName]
			if !ok {
				room = &Room{
					Name:    roomName,
					clients: make(map[*Client]bool),
				}
				h.rooms[roomName] = room
				log.Printf("Sala '%s' criada.", roomName)
			}

			// 3. Adiciona o cliente à nova sala.
			room.clients[client] = true
			// 4. Atualiza o estado do cliente.
			client.Room = room
			log.Printf("Cliente '%s' entrou na sala '%s'", client.Username, room.Name)

		case message := <-h.broadcast:
			if room, ok := h.rooms[message.RoomName]; ok {
				for client := range room.clients {
					client.send <- message.Body
				}
			}
		}
	}
}
