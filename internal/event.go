//Vamos criar structs para os payloads dos nossos eventos
package internal

import "encoding/json"

// Event é a estrutura base...
type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SendMessagePayload é o payload para o evento "send_message".
type SendMessagePayload struct {
	Content string `json:"content"`
	Room    string `json:"room"`
}

// JoinRoomPayload é o payload para o evento "join_room".
type JoinRoomPayload struct {
	Room string `json:"room"`
}
