package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
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

var jwtSecret = []byte("chave-secreta-super-dificil-de-adivinhar")

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var reqUser User
	if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
		http.Error(w, "Corpo da requisição inválido", http.StatusBadRequest)
		return
	}

	// 1. Busca o usuário no banco pelo nome de usuário
	storedUser, err := h.Store.GetUserByUsername(r.Context(), reqUser.Username)
	if err != nil {
		// Erro genérico para não informar se o usuário existe ou não (segurança).
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// 2. Compara a senha enviada com o hash salvo no banco
	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(reqUser.Password)); err != nil {
		// A senha não bate. Retorna o mesmo erro genérico.
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// 3. Se as credenciais estiverem corretas, gera o token JWT
	claims := jwt.MapClaims{
		"sub": storedUser.ID,                         // "subject", geralmente o ID do usuário
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Expira em 24 horas
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}

	log.Printf("Usuário '%s' logado com sucesso.", storedUser.Username)

	// 4. Retorna o token para o cliente
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}

// AuthMiddleware protege as rotas que exigem autenticação.
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Para WebSockets, é comum passar o token como um parâmetro de query.
		tokenString := r.URL.Query().Get("token")
		if tokenString == "" {
			http.Error(w, "Token não fornecido", http.StatusUnauthorized)
			return
		}

		// Valida o token.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verifica se o método de assinatura é o esperado.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("método de assinatura inesperado")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		// Se o token for válido, extrai o ID do usuário (claim "sub").
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if subject, ok := claims["sub"].(float64); ok {
				// Adiciona o ID do usuário ao contexto da requisição.
				ctx := context.WithValue(r.Context(), "userID", int(subject))
				// Chama o próximo handler na cadeia, passando o novo contexto.
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		http.Error(w, "Token inválido", http.StatusUnauthorized)
	})
}
