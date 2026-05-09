package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Event represents a single outbox record.
type Event struct {
	ID            uuid.UUID
	AggregateID   uuid.UUID
	AggregateType string
	EventType     string
	Payload       []byte
	Status        string
	Retries       int
	CreatedAt     time.Time
	ProcessedAt   *time.Time
}

// Repository defines the interface for outbox storage.
type Repository interface {
	SaveEvent(ctx context.Context, event Event) error
	MarkProcessed(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID) error
}

type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new outbox repository backed by Postgres.
func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) SaveEvent(ctx context.Context, event Event) error {
	query := `
		INSERT INTO outbox_events (id, aggregate_id, aggregate_type, event_type, payload, status, retries)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	// Ensure payload is valid JSON
	if !json.Valid(event.Payload) {
		return fmt.Errorf("invalid json payload")
	}

	_, err := r.db.ExecContext(ctx, query,
		event.ID,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.Payload,
		event.Status,
		event.Retries,
	)

	if err != nil {
		return fmt.Errorf("failed to insert outbox event: %w", err)
	}

	return nil
}

func (r *postgresRepository) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE outbox_events
		SET status = 'PROCESSED', processed_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark outbox event as processed: %w", err)
	}
	return nil
}

func (r *postgresRepository) MarkFailed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE outbox_events
		SET status = 'FAILED', processed_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark outbox event as failed: %w", err)
	}
	return nil
}
