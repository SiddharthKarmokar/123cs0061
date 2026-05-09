package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Config holds DB connection info.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// Connect establishes a connection to PostgreSQL and verifies it.
func Connect(ctx context.Context, cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
