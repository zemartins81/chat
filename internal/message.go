package internal

import "time"

// MessageBroadcast é usada para broadcast interno no hub
type MessageBroadcast struct {
	RoomName string
	Body     []byte
}

// Message é o modelo completo da mensagem para o banco
type Message struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	RoomName  string    `json:"room_name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
