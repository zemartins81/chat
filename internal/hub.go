// internal/hub.go

package internal

// Hub mantém o conjunto de clientes ativos e transmite mensagens para os clientes.
type Hub struct {
	// Clientes registrados. A chave é um ponteiro para o Cliente, o valor é booleano.
	clients map[*Client]bool

	// Canal de broadcast. Note que ele lida com []byte.
	// Ele não se importa se o conteúdo é texto puro ou JSON.
	broadcast chan []byte

	// Canal para registrar solicitações de clientes.
	register chan *Client

	// Canal para cancelar o registro de solicitações de clientes.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	// Este é o loop principal do Hub. Ele roda em sua própria goroutine.
	for {
		// O `select` espera que uma das operações de canal esteja pronta.
		select {
		case client := <-h.register:
			// Um novo cliente quer se registrar. Adicionamos ao nosso mapa.
			h.clients[client] = true
		case client := <-h.unregister:
			// Um cliente quer sair. Verificamos se ele existe, o removemos e fechamos seu canal.
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			// Recebemos uma mensagem para transmitir.
			// Percorremos todos os clientes conectados e enviamos a mensagem para o canal `send` deles.
			for client := range h.clients {
				select {
				case client.send <- message:
					// A mensagem foi enviada com sucesso.
				default:
					// Se o canal `send` estiver bloqueado, o cliente está lento ou morto.
					// O removemos para evitar bloqueios futuros.
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
