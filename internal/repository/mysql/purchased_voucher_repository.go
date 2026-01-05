package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/voucher-payment-service/internal/domain"
)

// PurchasedVoucherRepository implements repository.PurchasedVoucherRepository for MySQL
type PurchasedVoucherRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// NewPurchasedVoucherRepository creates a new purchased voucher repository
func NewPurchasedVoucherRepository(db *sql.DB) *PurchasedVoucherRepository {
	return &PurchasedVoucherRepository{db: db}
}

// WithTx returns a new repository with a transaction
func (r *PurchasedVoucherRepository) WithTx(tx *sql.Tx) *PurchasedVoucherRepository {
	return &PurchasedVoucherRepository{db: r.db, tx: tx}
}

// getDB returns the appropriate database connection
func (r *PurchasedVoucherRepository) getDB() interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create creates a new purchased voucher
func (r *PurchasedVoucherRepository) Create(ctx context.Context, pv *domain.PurchasedVoucher) error {
	query := `
		INSERT INTO purchased_vouchers (
			id, user_id, voucher_id, transaction_id, voucher_code, pin,
			status, purchase_price, valid_until, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err := r.getDB().ExecContext(ctx, query,
		pv.ID,
		pv.UserID,
		pv.VoucherID,
		pv.TransactionID,
		pv.VoucherCode,
		pv.Pin,
		pv.Status,
		pv.PurchasePrice,
		pv.ValidUntil,
	)

	if err != nil {
		return fmt.Errorf("failed to create purchased voucher: %w", err)
	}

	return nil
}

// CreateBatch creates multiple purchased vouchers in a batch
func (r *PurchasedVoucherRepository) CreateBatch(ctx context.Context, vouchers []*domain.PurchasedVoucher) error {
	if len(vouchers) == 0 {
		return nil
	}

	// Build batch insert query
	valueStrings := make([]string, 0, len(vouchers))
	valueArgs := make([]interface{}, 0, len(vouchers)*9)

	for _, pv := range vouchers {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())")
		valueArgs = append(valueArgs,
			pv.ID,
			pv.UserID,
			pv.VoucherID,
			pv.TransactionID,
			pv.VoucherCode,
			pv.Pin,
			pv.Status,
			pv.PurchasePrice,
			pv.ValidUntil,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO purchased_vouchers (
			id, user_id, voucher_id, transaction_id, voucher_code, pin,
			status, purchase_price, valid_until, created_at, updated_at
		) VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := r.getDB().ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to create purchased vouchers batch: %w", err)
	}

	return nil
}

// GetByID retrieves a purchased voucher by ID
func (r *PurchasedVoucherRepository) GetByID(ctx context.Context, id string) (*domain.PurchasedVoucher, error) {
	query := `
		SELECT id, user_id, voucher_id, transaction_id, voucher_code, pin,
		       status, purchase_price, valid_until, used_at, created_at, updated_at
		FROM purchased_vouchers
		WHERE id = ?
	`

	pv := &domain.PurchasedVoucher{}
	var usedAt sql.NullTime

	err := r.getDB().QueryRowContext(ctx, query, id).Scan(
		&pv.ID,
		&pv.UserID,
		&pv.VoucherID,
		&pv.TransactionID,
		&pv.VoucherCode,
		&pv.Pin,
		&pv.Status,
		&pv.PurchasePrice,
		&pv.ValidUntil,
		&usedAt,
		&pv.CreatedAt,
		&pv.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get purchased voucher: %w", err)
	}

	if usedAt.Valid {
		pv.UsedAt = &usedAt.Time
	}

	return pv, nil
}

// ListByUserID lists purchased vouchers for a user
func (r *PurchasedVoucherRepository) ListByUserID(ctx context.Context, userID, status string, page, pageSize int32) ([]*domain.PurchasedVoucher, int32, error) {
	// Build query based on filters
	baseQuery := "FROM purchased_vouchers WHERE user_id = ?"
	args := []interface{}{userID}

	if status != "" && status != "all" {
		baseQuery += " AND status = ?"
		args = append(args, status)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalCount int32
	err := r.getDB().QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count purchased vouchers: %w", err)
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
		SELECT id, user_id, voucher_id, transaction_id, voucher_code, pin,
		       status, purchase_price, valid_until, used_at, created_at, updated_at
	` + baseQuery + " ORDER BY created_at DESC LIMIT ? OFFSET ?"

	args = append(args, pageSize, offset)

	rows, err := r.getDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list purchased vouchers: %w", err)
	}
	defer rows.Close()

	vouchers := make([]*domain.PurchasedVoucher, 0)
	for rows.Next() {
		pv := &domain.PurchasedVoucher{}
		var usedAt sql.NullTime

		err := rows.Scan(
			&pv.ID,
			&pv.UserID,
			&pv.VoucherID,
			&pv.TransactionID,
			&pv.VoucherCode,
			&pv.Pin,
			&pv.Status,
			&pv.PurchasePrice,
			&pv.ValidUntil,
			&usedAt,
			&pv.CreatedAt,
			&pv.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan purchased voucher: %w", err)
		}

		if usedAt.Valid {
			pv.UsedAt = &usedAt.Time
		}

		vouchers = append(vouchers, pv)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating purchased vouchers: %w", err)
	}

	return vouchers, totalCount, nil
}

// UpdateStatus updates voucher status
func (r *PurchasedVoucherRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE purchased_vouchers
		SET status = ?,
		    updated_at = NOW()
		WHERE id = ?
	`

	result, err := r.getDB().ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update purchased voucher status: %w", err)
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

