package config

import (
	"log"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres PostgresConfig
	HTTP     HTTPConfig
	Telegram TelegramConfig
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST"`
	Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_USER"`
	DB       string `env:"POSTGRES_DB"`
	Password string `env:"POSTGRES_PASSWORD"`
	SSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`
	Timeout  string `env:"POSTGRES_TIMEOUT" envDefault:"5s"`
}

type HTTPConfig struct {
	Port string `env:"HTTP_PORT" envDefault:"8080"`
}

type TelegramConfig struct {
	Token       string        `env:"TELEGRAM_TOKEN"`
	PollTimeout time.Duration `env:"TELEGRAM_POLL_TIMEOUT" envDefault:"10s"`
}

var (
	instance *Config
	once     sync.Once
)

func MustLoad() *Config {
	once.Do(func() {
		var cfg Config
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("Error loading environment variables: %v", err)
		}
		instance = &cfg
	})
	return instance
}
