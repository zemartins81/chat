package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

// O upgrader é responsável por transformar uma conexão HTTP em uma conexão WebSocket.
// Ele define o tamanho dos buffers de leitura e escrita e a função CheckOrigin para permitir conexões de qualquer origem.
// Isso é útil para desenvolvimento, mas em produção você deve restringir as origens permitidas
// para evitar problemas de segurança, como ataques de Cross-Site WebSocket Hijacking (CSWH).
var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Permite conexões de qualquer origem
	}
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Servidor de chat rodando!")
	})

	log.Println("Servidor iniciado na porta 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}

}
