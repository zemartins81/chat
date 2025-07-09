package internal

import (
	"context"
	"log"

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
