package repository

import (
	"context"
	"github.com/voucher-payment-service/internal/domain"
)

// VoucherRepository defines methods for voucher data access
type VoucherRepository interface {
	// GetByID retrieves a voucher by ID
	GetByID(ctx context.Context, id string) (*domain.Voucher, error)
	
	// Search searches for vouchers based on criteria
	Search(ctx context.Context, params domain.SearchVouchersParams) ([]*domain.Voucher, int32, error)
	
	// UpdateStock updates voucher stock (with optimistic locking)
	UpdateStock(ctx context.Context, voucherID string, quantity int32) error
	
	// List retrieves all active vouchers
	List(ctx context.Context, page, pageSize int32) ([]*domain.Voucher, int32, error)
}

// WalletRepository defines methods for wallet data access
type WalletRepository interface {
	// GetByUserID retrieves a wallet by user ID
	GetByUserID(ctx context.Context, userID string) (*domain.Wallet, error)
	
	// UpdateBalance updates wallet balance (atomic operation)
	UpdateBalance(ctx context.Context, userID string, amount float64, txType string) (*domain.Wallet, error)
	
	// LockWallet locks a wallet for transactions
	LockWallet(ctx context.Context, userID string) error
	
	// UnlockWallet unlocks a wallet
	UnlockWallet(ctx context.Context, userID string) error
}

// TransactionRepository defines methods for transaction data access
type TransactionRepository interface {
	// Create creates a new transaction
	Create(ctx context.Context, tx *domain.Transaction) error
	
	// GetByID retrieves a transaction by ID
	GetByID(ctx context.Context, id string) (*domain.Transaction, error)
	
	// ListByUserID lists transactions for a user
	ListByUserID(ctx context.Context, userID string, txType string, page, pageSize int32) ([]*domain.Transaction, int32, error)
	
	// UpdateStatus updates transaction status
	UpdateStatus(ctx context.Context, id, status string) error
}

// PurchasedVoucherRepository defines methods for purchased voucher data access
type PurchasedVoucherRepository interface {
	// Create creates a new purchased voucher
	Create(ctx context.Context, pv *domain.PurchasedVoucher) error
	
	// CreateBatch creates multiple purchased vouchers in a batch
	CreateBatch(ctx context.Context, vouchers []*domain.PurchasedVoucher) error
	
	// GetByID retrieves a purchased voucher by ID
	GetByID(ctx context.Context, id string) (*domain.PurchasedVoucher, error)
	
	// ListByUserID lists purchased vouchers for a user
	ListByUserID(ctx context.Context, userID, status string, page, pageSize int32) ([]*domain.PurchasedVoucher, int32, error)
	
	// UpdateStatus updates voucher status
	UpdateStatus(ctx context.Context, id, status string) error
}

// UserRepository defines methods for user data access
type UserRepository interface {
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id string) (*domain.User, error)
	
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	
	// Create creates a new user
	Create(ctx context.Context, user *domain.User) error
}

// PaymentEventRepository defines methods for payment event data access
type PaymentEventRepository interface {
	// Create creates a new payment event
	Create(ctx context.Context, event *domain.PaymentEvent) error
	
	// GetByID retrieves a payment event by ID
	GetByID(ctx context.Context, id string) (*domain.PaymentEvent, error)
	
	// UpdateStatus updates event status
	UpdateStatus(ctx context.Context, id, status string) error
	
	// ListPending lists pending events for processing
	ListPending(ctx context.Context, limit int32) ([]*domain.PaymentEvent, error)
}

// UnitOfWork defines a transaction boundary
type UnitOfWork interface {
	// Begin starts a new transaction
	Begin(ctx context.Context) error
	
	// Commit commits the transaction
	Commit() error
	
	// Rollback rolls back the transaction
	Rollback() error
	
	// Vouchers returns the voucher repository
	Vouchers() VoucherRepository
	
	// Wallets returns the wallet repository
	Wallets() WalletRepository
	
	// Transactions returns the transaction repository
	Transactions() TransactionRepository
	
	// PurchasedVouchers returns the purchased voucher repository
	PurchasedVouchers() PurchasedVoucherRepository
	
	// Users returns the user repository
	Users() UserRepository
	
	// PaymentEvents returns the payment event repository
	PaymentEvents() PaymentEventRepository
}

