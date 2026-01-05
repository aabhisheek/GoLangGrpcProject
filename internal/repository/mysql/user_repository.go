package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voucher-payment-service/internal/domain"
)

// UserRepository implements repository.UserRepository for MySQL
type UserRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// WithTx returns a new repository with a transaction
func (r *UserRepository) WithTx(tx *sql.Tx) *UserRepository {
	return &UserRepository{db: r.db, tx: tx}
}

// getDB returns the appropriate database connection
func (r *UserRepository) getDB() interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, full_name, phone_number, status, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	user := &domain.User{}
	var phoneNumber sql.NullString

	err := r.getDB().QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&phoneNumber,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if phoneNumber.Valid {
		user.PhoneNumber = phoneNumber.String
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, full_name, phone_number, status, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	user := &domain.User{}
	var phoneNumber sql.NullString

	err := r.getDB().QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&phoneNumber,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if phoneNumber.Valid {
		user.PhoneNumber = phoneNumber.String
	}

	return user, nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, full_name, phone_number, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`

	var phoneNumber interface{}
	if user.PhoneNumber != "" {
		phoneNumber = user.PhoneNumber
	}

	_, err := r.getDB().ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.FullName,
		phoneNumber,
		user.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

