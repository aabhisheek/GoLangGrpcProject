package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/voucher-payment-service/internal/domain"
)

// VoucherRepository implements repository.VoucherRepository for MySQL
type VoucherRepository struct {
	db *sql.DB
	tx *sql.Tx
}

// NewVoucherRepository creates a new voucher repository
func NewVoucherRepository(db *sql.DB) *VoucherRepository {
	return &VoucherRepository{db: db}
}

// WithTx returns a new repository with a transaction
func (r *VoucherRepository) WithTx(tx *sql.Tx) *VoucherRepository {
	return &VoucherRepository{db: r.db, tx: tx}
}

// getDB returns the appropriate database connection
func (r *VoucherRepository) getDB() interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// GetByID retrieves a voucher by ID
func (r *VoucherRepository) GetByID(ctx context.Context, id string) (*domain.Voucher, error) {
	query := `
		SELECT id, name, description, brand, category, face_value, discount_percentage,
		       selling_price, stock_quantity, is_active, valid_from, valid_until,
		       terms_and_conditions, created_at, updated_at
		FROM vouchers
		WHERE id = ?
	`

	voucher := &domain.Voucher{}
	err := r.getDB().QueryRowContext(ctx, query, id).Scan(
		&voucher.ID,
		&voucher.Name,
		&voucher.Description,
		&voucher.Brand,
		&voucher.Category,
		&voucher.FaceValue,
		&voucher.DiscountPercentage,
		&voucher.SellingPrice,
		&voucher.StockQuantity,
		&voucher.IsActive,
		&voucher.ValidFrom,
		&voucher.ValidUntil,
		&voucher.TermsAndConditions,
		&voucher.CreatedAt,
		&voucher.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get voucher: %w", err)
	}

	return voucher, nil
}

// Search searches for vouchers based on criteria
func (r *VoucherRepository) Search(ctx context.Context, params domain.SearchVouchersParams) ([]*domain.Voucher, int32, error) {
	// Build query with filters
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "is_active = ?")
	args = append(args, true)

	if params.Query != "" {
		conditions = append(conditions, "MATCH(name, description, brand) AGAINST(? IN NATURAL LANGUAGE MODE)")
		args = append(args, params.Query)
	}

	if params.Category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, params.Category)
	}

	if params.Brand != "" {
		conditions = append(conditions, "brand = ?")
		args = append(args, params.Brand)
	}

	if params.MinPrice > 0 {
		conditions = append(conditions, "selling_price >= ?")
		args = append(args, params.MinPrice)
	}

	if params.MaxPrice > 0 {
		conditions = append(conditions, "selling_price <= ?")
		args = append(args, params.MaxPrice)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM vouchers WHERE %s", whereClause)
	var totalCount int32
	err := r.getDB().QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count vouchers: %w", err)
	}

	// Get paginated results
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
		SELECT id, name, description, brand, category, face_value, discount_percentage,
		       selling_price, stock_quantity, is_active, valid_from, valid_until,
		       terms_and_conditions, created_at, updated_at
		FROM vouchers
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, pageSize, offset)

	rows, err := r.getDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search vouchers: %w", err)
	}
	defer rows.Close()

	vouchers := make([]*domain.Voucher, 0)
	for rows.Next() {
		voucher := &domain.Voucher{}
		err := rows.Scan(
			&voucher.ID,
			&voucher.Name,
			&voucher.Description,
			&voucher.Brand,
			&voucher.Category,
			&voucher.FaceValue,
			&voucher.DiscountPercentage,
			&voucher.SellingPrice,
			&voucher.StockQuantity,
			&voucher.IsActive,
			&voucher.ValidFrom,
			&voucher.ValidUntil,
			&voucher.TermsAndConditions,
			&voucher.CreatedAt,
			&voucher.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan voucher: %w", err)
		}
		vouchers = append(vouchers, voucher)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating vouchers: %w", err)
	}

	return vouchers, totalCount, nil
}

// UpdateStock updates voucher stock with optimistic locking
func (r *VoucherRepository) UpdateStock(ctx context.Context, voucherID string, quantity int32) error {
	query := `
		UPDATE vouchers
		SET stock_quantity = stock_quantity - ?,
		    updated_at = NOW()
		WHERE id = ? AND stock_quantity >= ? AND is_active = true
	`

	result, err := r.getDB().ExecContext(ctx, query, quantity, voucherID, quantity)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVoucherOutOfStock
	}

	return nil
}

// List retrieves all active vouchers
func (r *VoucherRepository) List(ctx context.Context, page, pageSize int32) ([]*domain.Voucher, int32, error) {
	params := domain.SearchVouchersParams{
		Page:     page,
		PageSize: pageSize,
	}
	return r.Search(ctx, params)
}

