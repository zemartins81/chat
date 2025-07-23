package internal

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

// NewStore cria uma nova conexão com o banco e retorna uma Store.
func NewStore(connString string) (*Store, error) {
	db, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, err
	}

	log.Println("Conexão com o PostgreSQL estabelecida com sucesso!")
	return &Store{db: db}, nil
}

// Init inicializa o banco de dados, criando as tabelas necessárias.
func (s *Store) Init() error {
	// Query para criar a tabela de usuários (já existente)
	createUsersTableQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(50) UNIQUE NOT NULL,
        password VARCHAR(255) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`
	_, err := s.db.Exec(context.Background(), createUsersTableQuery)
	if err != nil {
		return err
	}

	// NOVA QUERY: para criar a tabela de salas
	createRoomsTableQuery := `
	CREATE TABLE IF NOT EXISTS rooms (
		id SERIAL PRIMARY KEY,
		name VARCHAR(50) UNIQUE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = s.db.Exec(context.Background(), createRoomsTableQuery)
	if err != nil {
		return err
	}

	createMessagesTableQuery := `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		username VARCHAR(50) NOT NULL,
		room_name VARCHAR(50) NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = s.db.Exec(context.Background(), createMessagesTableQuery)
	if err != nil {
		return err
	}

	log.Println("Tabelas 'users', 'rooms' e 'messages' verificadas/criadas com sucesso.")
	return nil
}

func (s *Store) CreateUser(ctx context.Context, user *User) error {
	// Define a query SQL para inserir um novo usuário na tabela 'users'.
	// A query inclui os campos 'username' e 'password' e retorna o 'id' e 'created_at' do usuário recém-criado.
	query := `
        INSERT INTO users (username, password)
        VALUES ($1, $2)
        RETURNING id, created_at;`

	// Executa a query no banco de dados usando o banco de dados 's.db'.
	// O 'ctx' é um contexto que permite cancelar a operação se necessário.
	// 'user.Username' e 'user.Password' são os valores a serem inseridos na query.
	if err := s.db.QueryRow(ctx, query, user.Username, user.Password).Scan(&user.ID, &user.CreatedAt); err != nil {
		// Verifica se o erro é um erro específico do PostgreSQL (pgErr).
		var pgErr *pgconn.PgError
		// Verifica se o erro é um erro do PostgreSQL e se o código de erro é "23505".
		// O código de erro "23505" indica que o nome do usuário já existe na tabela.
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Retorna um erro personalizado indicando que o nome do usuário já existe.
			return errors.New("nome do usuário já existe")
		}
		// Retorna o erro original se não for um erro de nome de usuário duplicado.
		return err
	}
	// Retorna nil se a inserção for bem-sucedida.
	return nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User

	query := `SELECT id, username, password, created_at FROM users WHERE username = $1`

	err := s.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	return &user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int) (*User, error) {
	var user User
	query := `SELECT id, username, created_at FROM users WHERE id = $1`
	err := s.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}
	return &user, nil
}

func (s *Store) SaveMessage(ctx context.Context, message *Message) error {
	query := `
		INSERT INTO messages (user_id, username, room_name, content)
		VALUES ($1, $2, $3, $4)
		ReTURNING id, created_at;`

	err := s.db.QueryRow(ctx, query,
		message.UserID,
		message.Username,
		message.RoomName,
		message.Content).Scan(&message.ID, &message.CreatedAt)

	return err
}

// GetRoomMessages busca as últimas mensagens de uma sala
func (s *Store) GetRoomMessages(ctx context.Context, roomName string, limit int) ([]*Message, error) {
	query := `
        SELECT id, user_id, username, room_name, content, created_at 
        FROM messages 
        WHERE room_name = $1 
        ORDER BY created_at DESC 
        LIMIT $2;`

	rows, err := s.db.Query(ctx, query, roomName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.Username,
			&msg.RoomName, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	// Inverte a ordem para ficar cronológica (mais antiga primeiro)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
