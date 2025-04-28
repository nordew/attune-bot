package db

import (
	"attune/internal/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxRetries = 5
)

var (
	retryDelay = 5 * time.Second
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

	log.Printf("Connecting to PostgreSQL database at %s", dsn)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("unable to parse connection string: %v", err)
	}

	if pgConfig.Timeout != "" {
		timeout, err := time.ParseDuration(pgConfig.Timeout)
		if err != nil {
			log.Printf("invalid timeout value: %v, using default", err)
		} else {
			poolConfig.ConnConfig.ConnectTimeout = timeout
		}
	}

	var pool *pgxpool.Pool

	for i := range make([]struct{}, maxRetries) {
		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			if err := pool.Ping(ctx); err == nil {
				log.Println("connected to PostgreSQL database")
				return pool, dsn
			}

			pool.Close()
			log.Printf("failed to ping database: %v, retrying...", err)
		} else {
			log.Printf("attempt %d/%d: unable to connect to database: %v", i+1, maxRetries, err)
		}

		if i < maxRetries-1 {
			log.Printf("retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	log.Fatalf("failed to connect to database after %d attempts", maxRetries)
	return nil, ""
}
