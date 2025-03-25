package continuous

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Connection struct {
	db *pgxpool.Pool
}

func GetConnection(cfg *Configuration) Connection {
	if cfg.db != nil {
		return cfg.Connection
	}
	log.Println("creating new DB connection")
	ctx := context.Background()
	cn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", cfg.DbUser, cfg.DbPass, cfg.DbAddr, cfg.DbName)
	db, err := pgxpool.New(ctx, cn)
	if err != nil {
		panic("db connection failed")
	}
	connection := Connection{db: db}
	cfg.Connection = connection
	return connection
}
