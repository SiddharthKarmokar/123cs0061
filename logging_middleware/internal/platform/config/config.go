package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config represents the entire application configuration loaded from environment variables.
type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	GRPC     GRPCConfig
	Postgres PostgresConfig
	Mongo    MongoConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Logger   LoggerConfig
}

type AppConfig struct {
	Env string `env:"APP_ENV" env-default:"development"`
}

type HTTPConfig struct {
	Port string `env:"HTTP_PORT" env-default:"8080"`
}

type GRPCConfig struct {
	Port string `env:"GRPC_PORT" env-default:"9090"`
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" env-default:"postgres"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DB       string `env:"POSTGRES_DB" env-required:"true"`
}

type MongoConfig struct {
	Host     string `env:"MONGO_HOST" env-default:"mongodb"`
	Port     string `env:"MONGO_PORT" env-default:"27017"`
	User     string `env:"MONGO_USER" env-required:"true"`
	Password string `env:"MONGO_PASSWORD" env-required:"true"`
	DB       string `env:"MONGO_DB" env-required:"true"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" env-default:"redis"`
	Port     string `env:"REDIS_PORT" env-default:"6379"`
	Password string `env:"REDIS_PASSWORD"`
}

type AuthConfig struct {
	BaseURL      string `env:"AUTH_BASE_URL" env-required:"true"`
	Email        string `env:"AUTH_EMAIL" env-required:"true"`
	Name         string `env:"AUTH_NAME" env-required:"true"`
	RollNo       string `env:"AUTH_ROLL_NO" env-required:"true"`
	AccessCode   string `env:"AUTH_ACCESS_CODE" env-required:"true"`
	ClientID     string `env:"AUTH_CLIENT_ID" env-required:"true"`
	ClientSecret string `env:"AUTH_CLIENT_SECRET" env-required:"true"`
}

type LoggerConfig struct {
	MaxWorkers int `env:"LOGGER_MAX_WORKERS" env-default:"5"`
	QueueSize  int `env:"LOGGER_QUEUE_SIZE" env-default:"1000"`
	MaxRetries int `env:"LOGGER_MAX_RETRIES" env-default:"3"`
}

// Load reads configuration from environment variables and an optional .env file.
func Load() (*Config, error) {
	var cfg Config

	// First try to load from .env file if it exists, but don't fail if it doesn't
	_ = cleanenv.ReadConfig(".env", &cfg)

	// Always parse environment variables (overrides .env)
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return &cfg, nil
}
