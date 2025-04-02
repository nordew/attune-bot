package app

import (
	"attune/internal/api"
	"attune/internal/api/telegram"
	"attune/internal/config"
	"attune/internal/service"
	"attune/internal/storage"
	"attune/pkg/cache"
	"attune/pkg/db"
	"attune/pkg/logger"
	"attune/pkg/transactor"
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/pressly/goose"
)

func MustRun() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()
	pgConn, dsn := db.MustConnect(ctx, cfg.Postgres)

	migrate(dsn)

	slog := logger.NewSLogger()
	storages := storage.NewStorages(pgConn)

	customCache := cache.NewCache()
	pgxTx := transactor.NewPgxTransactor(pgConn)

	apiCh := make(chan api.Trigger)
	stopCh := make(chan struct{})

	focusSessionManager := service.NewFocusSessionManager(storages, customCache, apiCh)
	services := service.NewServices(storages, focusSessionManager, pgxTx, slog, customCache)

	telegramAPI := telegram.NewTelegramAPI(cfg.Telegram.Token, "base_url", cfg.Telegram.PollTimeout, *services, slog, customCache, apiCh)

	go func() {
		cacheCfg := cache.StartCacheWorkerConfig{
			Interval: time.Minute,
			StopCh:   stopCh,
			Cache:    customCache,
		}

		cache.StartCacheWorker(context.Background(), cacheCfg)
	}()

	go func() {
		if err := telegramAPI.Start(ctx); err != nil {
			log.Fatalf("failed to start telegram API: %v", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signalChan
	log.Printf("signal received: %v", sig)

	_, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	pgConn.Close()
	log.Print("Postgres connection closed")

	log.Print("shutting down services")
}

func migrate(dsn string) {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open SQL connection: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set goose dialect: %v", err)
	}

	if err := goose.Up(sqlDB, "./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}
