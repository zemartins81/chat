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

// serveWs agora recebe o store para buscar dados do usuário.
func serveWs(hub *internal.Hub, store *internal.Store, w http.ResponseWriter, r *http.Request) {
	// Extrai o userID do contexto, que foi adicionado pelo middleware.
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Não foi possível obter o ID do usuário", http.StatusInternalServerError)
		return
	}

	// Busca os dados do usuário no banco.
	user, err := store.GetUserByID(r.Context(), userID) // Precisaremos criar este método!
	if err != nil {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Cria o cliente com as informações do usuário.
	client := internal.NewClient(hub, conn)
	client.UserID = user.ID
	client.Username = user.Username

	client.Register()
	log.Printf("Cliente conectado: %s (ID: %d)", client.Username, client.UserID)

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

	apiHandler := &internal.Handler{Store: store}

	http.HandleFunc("/register", apiHandler.RegisterHandler)
	http.HandleFunc("/login", apiHandler.LoginHandler)

	// Protegendo a rota /ws com o middleware.
	wsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, store, w, r)
	})
	http.Handle("/ws", apiHandler.AuthMiddleware(wsHandler))

	log.Println("Servidor iniciado na porta :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Erro ao iniciar o servidor: ", err)
	}
}
