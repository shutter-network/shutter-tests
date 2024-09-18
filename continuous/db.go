package continuous

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Connection struct {
	db *pgxpool.Pool
}

func NewConnection(cfg *Configuration) Connection {
	DbUser := "postgres"
	DbPass := "test"
	dbAddr := "localhost:5432"
	DbName := "shutter_metrics"

	ctx := context.Background()
	db, err := pgxpool.New(ctx, fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", DbUser, DbPass, dbAddr, DbName))
	if err != nil {
		panic("db connection failed")
	}
	connection := Connection{db: db}
	return connection
}
