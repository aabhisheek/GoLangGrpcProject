package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voucher-payment-service/internal/domain"
)

// WalletRepository implements repository.WalletRepository for MySQL
type WalletRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// NewWalletRepository creates a new wallet repository
func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// WithTx returns a new repository with a transaction
func (r *WalletRepository) WithTx(tx *sql.Tx) *WalletRepository {
	return &WalletRepository{db: r.db, tx: tx}
}

// getDB returns the appropriate database connection
func (r *WalletRepository) getDB() interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// GetByUserID retrieves a wallet by user ID with row locking
func (r *WalletRepository) GetByUserID(ctx context.Context, userID string) (*domain.Wallet, error) {
	query := `
		SELECT id, user_id, balance, currency, is_locked, created_at, updated_at
		FROM wallets
		WHERE user_id = ?
		FOR UPDATE
	`

	wallet := &domain.Wallet{}
	err := r.getDB().QueryRowContext(ctx, query, userID).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Balance,
		&wallet.Currency,
		&wallet.IsLocked,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	if wallet.IsLocked {
		return nil, domain.ErrWalletLocked
	}

	return wallet, nil
}

// UpdateBalance updates wallet balance atomically
func (r *WalletRepository) UpdateBalance(ctx context.Context, userID string, amount float64, txType string) (*domain.Wallet, error) {
	// First, get current balance with lock
	wallet, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var newBalance float64
	if txType == "credit" {
		newBalance = wallet.Balance + amount
	} else if txType == "debit" {
		if wallet.Balance < amount {
			return nil, domain.ErrInsufficientBalance
		}
		newBalance = wallet.Balance - amount
	} else {
		return nil, fmt.Errorf("invalid transaction type: %s", txType)
	}

	// Update balance
	query := `
		UPDATE wallets
		SET balance = ?,
		    updated_at = NOW()
		WHERE user_id = ? AND is_locked = false
	`

	result, err := r.getDB().ExecContext(ctx, query, newBalance, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, domain.ErrWalletLocked
	}

	wallet.Balance = newBalance
	return wallet, nil
}

// LockWallet locks a wallet for transactions
func (r *WalletRepository) LockWallet(ctx context.Context, userID string) error {
	query := `
		UPDATE wallets
		SET is_locked = true,
		    updated_at = NOW()
		WHERE user_id = ?
	`

	_, err := r.getDB().ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to lock wallet: %w", err)
	}

	return nil
}

// UnlockWallet unlocks a wallet
func (r *WalletRepository) UnlockWallet(ctx context.Context, userID string) error {
	query := `
		UPDATE wallets
		SET is_locked = false,
		    updated_at = NOW()
		WHERE user_id = ?
	`

	_, err := r.getDB().ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to unlock wallet: %w", err)
	}

	return nil
}

