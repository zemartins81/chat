package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// O upgrader é responsável por transformar uma conexão HTTP em uma conexão WebSocket.
// Ele definse o tamanho dos buffers de leitura e escrita e a função CheckOrigin para permitir conexões de qualquer origem.
// Isso é útil para desenvolvimento, mas em produção você deve restringir as origens permitidas
// para evitar problemas de segurança, como ataques de Cross-Site WebSocket Hijacking (CSWH).
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Permite conexões de qualquer origem
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// 1. Faz o "upgrade" da conexão HTTP para WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro ao fazer upgrade para WebSocket: %v", err)
		return
	}
	defer ws.Close()

	log.Println("Nova conexão WebSocket estabelecida")

	// 2. Loop infinito para ouvir mensagens do cliente
	for {
		messageType, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Erro ao ler a mensagem")
			break
		}

		log.Printf("recebido: %s", msg)

		// 3. Devolve a mensagem para o mesmo cliente (efeito "eco")
		err = ws.WriteMessage(messageType, msg)
		if err != nil {
			log.Println("Erro ao escrever a mensagem")
			break
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)

	log.Println("Servidor websocket iniciado na porta 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
