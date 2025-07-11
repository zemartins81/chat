package internal

type Message struct {
	Username string `json:"username,omitempty"`
	Content  string `json:"content,omitempty"`
}
