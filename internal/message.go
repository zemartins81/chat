package internal

// Message agora contém o corpo da mensagem e o nome da sala de destino.
type Message struct {
	RoomName string
	Body     []byte
}
