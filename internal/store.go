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
	// Vamos criar a tabela de usuários se ela não existir.
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(50) UNIQUE NOT NULL,
        password VARCHAR(255) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`

	_, err := s.db.Exec(context.Background(), createTableQuery)
	if err != nil {
		return err
	}

	log.Println("Tabela 'users' verificada/criada com sucesso.")
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
