package internal

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Store *Store
}

// RegisterHandler lida com as requisições de registro de novos usuários.
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var user User
	// Decodifica o JSON do corpo da requisição para a nossa struct User.
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Corpo da requisição inválido", http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" {
		http.Error(w, "Usuário e senha são obrigatórios", http.StatusBadRequest)
		return
	}

	// Gera o hash da senha do usuário com bcrypt.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Erro ao processar a senha", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	// Chama o método do store para criar o usuário no banco.
	if err := h.Store.CreateUser(r.Context(), &user); err != nil {
		if err.Error() == "nome de usuário já existe" {
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
		} else {
			http.Error(w, "Erro ao criar usuário", http.StatusInternalServerError)
		}
		return
	}

	// Remove a senha antes de retornar a resposta para o cliente.
	user.Password = ""

	log.Printf("Novo usuário registrado: %s", user.Username)

	// Responde com o status 201 Created e os dados do usuário em JSON.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
