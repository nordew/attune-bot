package db

import (
	"attune/internal/config"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MustConnect(ctx context.Context, pgConfig config.PostgresConfig) (*pgxpool.Pool, string) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		pgConfig.Host,
		pgConfig.Port,
		pgConfig.User,
		pgConfig.DB,
		pgConfig.Password,
		pgConfig.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("unable to parse connection string: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("unable to ping database: %v", err)
	}

	log.Println("connected to PostgreSQL database")

	return pool, dsn
}
