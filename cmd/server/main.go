package main

import (
	"log"
	"net/http"

	"chat/internal"

	"github.com/gorilla/websocket"
)

// O upgrader é responsável por transformar uma conexão HTTP em uma conexão WebSocket.
// Ele define o tamanho dos buffers de leitura e escrita e a função CheckOrigin para permitir conexões de qualquer origem.
// Isso é útil para desenvolvimento, mas em produção você deve restringir as origens permitidas
// para evitar problemas de segurança, como ataques de Cross-Site WebSocket Hijacking (CSWH).
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Permite conexões de qualquer origem
	},
}

func serveWs(hub *internal.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Cria um novo cliente
	client := internal.NewClient(hub, conn)
	// Registra o cliente no Hub
	client.Register()

	// Inicia as goroutines para ler e escrever mensagens para este cliente
	go client.WritePump()
	go client.ReadPump()
}

func main() {
	connString := "postgres://user:password@localhost:5432/chatdb?sslmode=disable"
	store, err := internal.NewStore(connString)
	if err != nil {
		log.Fatalf("Não foi possível conectar ao banco de dados: %v", err)
	}

	if err := store.Init(); err != nil {
		log.Fatalf("Não foi possível inicializar o banco de dados: %v", err)
	}

	hub := internal.NewHub()
	go hub.Run()

	// Cria uma instância do nosso Handler da API, injetando o store.
	apiHandler := &internal.Handler{Store: store}

	// Registra a nova rota de registro.
	http.HandleFunc("/register", apiHandler.RegisterHandler)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Println("Servidor iniciado na porta :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Erro ao iniciar o servidor: ", err)
	}
}
