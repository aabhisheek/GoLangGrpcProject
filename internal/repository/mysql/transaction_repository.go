package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voucher-payment-service/internal/domain"
)

// TransactionRepository implements repository.TransactionRepository for MySQL
type TransactionRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// WithTx returns a new repository with a transaction
func (r *TransactionRepository) WithTx(tx *sql.Tx) *TransactionRepository {
	return &TransactionRepository{db: r.db, tx: tx}
}

// getDB returns the appropriate database connection
func (r *TransactionRepository) getDB() interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create creates a new transaction
func (r *TransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, user_id, type, amount, balance_before, balance_after,
			description, reference_id, payment_method, status, metadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := r.getDB().ExecContext(ctx, query,
		transaction.ID,
		transaction.UserID,
		transaction.Type,
		transaction.Amount,
		transaction.BalanceBefore,
		transaction.BalanceAfter,
		transaction.Description,
		transaction.ReferenceID,
		transaction.PaymentMethod,
		transaction.Status,
		transaction.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a transaction by ID
func (r *TransactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, balance_before, balance_after,
		       description, reference_id, payment_method, status, metadata,
		       created_at, updated_at
		FROM transactions
		WHERE id = ?
	`

	transaction := &domain.Transaction{}
	var metadata sql.NullString

	err := r.getDB().QueryRowContext(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.Type,
		&transaction.Amount,
		&transaction.BalanceBefore,
		&transaction.BalanceAfter,
		&transaction.Description,
		&transaction.ReferenceID,
		&transaction.PaymentMethod,
		&transaction.Status,
		&metadata,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if metadata.Valid {
		transaction.Metadata = metadata.String
	}

	return transaction, nil
}

// ListByUserID lists transactions for a user
func (r *TransactionRepository) ListByUserID(ctx context.Context, userID string, txType string, page, pageSize int32) ([]*domain.Transaction, int32, error) {
	// Build query based on filters
	baseQuery := "FROM transactions WHERE user_id = ?"
	args := []interface{}{userID}

	if txType != "" && txType != "all" {
		baseQuery += " AND type = ?"
		args = append(args, txType)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalCount int32
	err := r.getDB().QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Get paginated results
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	query := `
		SELECT id, user_id, type, amount, balance_before, balance_after,
		       description, reference_id, payment_method, status, metadata,
		       created_at, updated_at
	` + baseQuery + " ORDER BY created_at DESC LIMIT ? OFFSET ?"

	args = append(args, pageSize, offset)

	rows, err := r.getDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]*domain.Transaction, 0)
	for rows.Next() {
		transaction := &domain.Transaction{}
		var metadata sql.NullString

		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.Type,
			&transaction.Amount,
			&transaction.BalanceBefore,
			&transaction.BalanceAfter,
			&transaction.Description,
			&transaction.ReferenceID,
			&transaction.PaymentMethod,
			&transaction.Status,
			&metadata,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}

		if metadata.Valid {
			transaction.Metadata = metadata.String
		}

		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, totalCount, nil
}

// UpdateStatus updates transaction status
func (r *TransactionRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE transactions
		SET status = ?,
		    updated_at = NOW()
		WHERE id = ?
	`

	result, err := r.getDB().ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
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

