package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	PhoneNumber string    `json:"phone_number"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Wallet represents a user's wallet
type Wallet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	IsLocked  bool      `json:"is_locked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Voucher represents an available voucher
type Voucher struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Brand              string    `json:"brand"`
	Category           string    `json:"category"`
	FaceValue          float64   `json:"face_value"`
	DiscountPercentage float64   `json:"discount_percentage"`
	SellingPrice       float64   `json:"selling_price"`
	StockQuantity      int32     `json:"stock_quantity"`
	IsActive           bool      `json:"is_active"`
	ValidFrom          time.Time `json:"valid_from"`
	ValidUntil         time.Time `json:"valid_until"`
	TermsAndConditions string    `json:"terms_and_conditions"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// PurchasedVoucher represents a voucher owned by a user
type PurchasedVoucher struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	VoucherID     string    `json:"voucher_id"`
	TransactionID string    `json:"transaction_id"`
	VoucherCode   string    `json:"voucher_code"`
	Pin           string    `json:"pin"`
	Status        string    `json:"status"`
	PurchasePrice float64   `json:"purchase_price"`
	ValidUntil    time.Time `json:"valid_until"`
	UsedAt        *time.Time `json:"used_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Transaction represents a wallet transaction
type Transaction struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Type          string    `json:"type"` // credit, debit
	Amount        float64   `json:"amount"`
	BalanceBefore float64   `json:"balance_before"`
	BalanceAfter  float64   `json:"balance_after"`
	Description   string    `json:"description"`
	ReferenceID   string    `json:"reference_id"`
	PaymentMethod string    `json:"payment_method"`
	Status        string    `json:"status"`
	Metadata      string    `json:"metadata,omitempty"` // JSON string
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PaymentEvent represents an event for message queue
type PaymentEvent struct {
	ID            string    `json:"id"`
	EventType     string    `json:"event_type"`
	UserID        string    `json:"user_id"`
	TransactionID string    `json:"transaction_id,omitempty"`
	Payload       string    `json:"payload"` // JSON string
	Status        string    `json:"status"`
	RetryCount    int       `json:"retry_count"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// VoucherPurchaseEvent represents a voucher purchase event
type VoucherPurchaseEvent struct {
	TransactionID string   `json:"transaction_id"`
	UserID        string   `json:"user_id"`
	VoucherID     string   `json:"voucher_id"`
	Quantity      int32    `json:"quantity"`
	TotalAmount   float64  `json:"total_amount"`
	VoucherCodes  []string `json:"voucher_codes"`
	Timestamp     int64    `json:"timestamp"`
}

// SearchVouchersParams parameters for searching vouchers
type SearchVouchersParams struct {
	Query    string
	Category string
	Brand    string
	MinPrice float64
	MaxPrice float64
	Page     int32
	PageSize int32
}

// PaginatedResult generic paginated result
type PaginatedResult struct {
	TotalCount int32
	Page       int32
	PageSize   int32
}

