package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voucher-payment-service/internal/domain"
)

// PaymentEventRepository implements repository.PaymentEventRepository for MySQL
type PaymentEventRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// NewPaymentEventRepository creates a new payment event repository
func NewPaymentEventRepository(db *sql.DB) *PaymentEventRepository {
	return &PaymentEventRepository{db: db}
}

// WithTx returns a new repository with a transaction
func (r *PaymentEventRepository) WithTx(tx *sql.Tx) *PaymentEventRepository {
	return &PaymentEventRepository{db: r.db, tx: tx}
}

// getDB returns the appropriate database connection
func (r *PaymentEventRepository) getDB() interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create creates a new payment event
func (r *PaymentEventRepository) Create(ctx context.Context, event *domain.PaymentEvent) error {
	query := `
		INSERT INTO payment_events (
			id, event_type, user_id, transaction_id, payload, status,
			retry_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := r.getDB().ExecContext(ctx, query,
		event.ID,
		event.EventType,
		event.UserID,
		event.TransactionID,
		event.Payload,
		event.Status,
		event.RetryCount,
	)

	if err != nil {
		return fmt.Errorf("failed to create payment event: %w", err)
	}

	return nil
}

// GetByID retrieves a payment event by ID
func (r *PaymentEventRepository) GetByID(ctx context.Context, id string) (*domain.PaymentEvent, error) {
	query := `
		SELECT id, event_type, user_id, transaction_id, payload, status,
		       retry_count, processed_at, created_at, updated_at
		FROM payment_events
		WHERE id = ?
	`

	event := &domain.PaymentEvent{}
	var transactionID sql.NullString
	var processedTime sql.NullTime

	err := r.getDB().QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.EventType,
		&event.UserID,
		&transactionID,
		&event.Payload,
		&event.Status,
		&event.RetryCount,
		&processedTime,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment event: %w", err)
	}

	if transactionID.Valid {
		event.TransactionID = transactionID.String
	}
	if processedTime.Valid {
		event.ProcessedAt = &processedTime.Time
	}

	return event, nil
}

// UpdateStatus updates event status
func (r *PaymentEventRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE payment_events
		SET status = ?,
		    processed_at = CASE WHEN ? = 'processed' THEN NOW() ELSE processed_at END,
		    updated_at = NOW()
		WHERE id = ?
	`

	result, err := r.getDB().ExecContext(ctx, query, status, status, id)
	if err != nil {
		return fmt.Errorf("failed to update payment event status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ListPending lists pending events for processing
func (r *PaymentEventRepository) ListPending(ctx context.Context, limit int32) ([]*domain.PaymentEvent, error) {
	query := `
		SELECT id, event_type, user_id, transaction_id, payload, status,
		       retry_count, processed_at, created_at, updated_at
		FROM payment_events
		WHERE status = 'pending' AND retry_count < 5
		ORDER BY created_at ASC
		LIMIT ?
	`

	rows, err := r.getDB().QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending events: %w", err)
	}
	defer rows.Close()

	events := make([]*domain.PaymentEvent, 0)
	for rows.Next() {
		event := &domain.PaymentEvent{}
		var transactionID sql.NullString
		var processedTime sql.NullTime

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.UserID,
			&transactionID,
			&event.Payload,
			&event.Status,
			&event.RetryCount,
			&processedTime,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment event: %w", err)
		}

		if transactionID.Valid {
			event.TransactionID = transactionID.String
		}
		if processedTime.Valid {
			event.ProcessedAt = &processedTime.Time
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment events: %w", err)
	}

	return events, nil
}

