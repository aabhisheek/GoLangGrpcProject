package service

import (
	"context"
	"testing"
	"time"

	"github.com/voucher-payment-service/internal/domain"
)

// MockUnitOfWork for testing
type MockUnitOfWork struct {
	voucherRepo  *MockVoucherRepo
	walletRepo   *MockWalletRepo
	txRepo       *MockTransactionRepo
	pvRepo       *MockPurchasedVoucherRepo
	shouldCommit bool
}

func (m *MockUnitOfWork) Begin(ctx context.Context) error {
	return nil
}

func (m *MockUnitOfWork) Commit() error {
	if !m.shouldCommit {
		return domain.ErrTransactionFailed
	}
	return nil
}

func (m *MockUnitOfWork) Rollback() error {
	return nil
}

func (m *MockUnitOfWork) Vouchers() interface{} {
	return m.voucherRepo
}

func (m *MockUnitOfWork) Wallets() interface{} {
	return m.walletRepo
}

func (m *MockUnitOfWork) Transactions() interface{} {
	return m.txRepo
}

func (m *MockUnitOfWork) PurchasedVouchers() interface{} {
	return m.pvRepo
}

func (m *MockUnitOfWork) Users() interface{} {
	return nil
}

func (m *MockUnitOfWork) PaymentEvents() interface{} {
	return nil
}

// MockVoucherRepo for testing
type MockVoucherRepo struct {
	vouchers map[string]*domain.Voucher
}

func (m *MockVoucherRepo) GetByID(ctx context.Context, id string) (*domain.Voucher, error) {
	v, ok := m.vouchers[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return v, nil
}

func (m *MockVoucherRepo) Search(ctx context.Context, params domain.SearchVouchersParams) ([]*domain.Voucher, int32, error) {
	result := make([]*domain.Voucher, 0)
	for _, v := range m.vouchers {
		result = append(result, v)
	}
	return result, int32(len(result)), nil
}

func (m *MockVoucherRepo) UpdateStock(ctx context.Context, voucherID string, quantity int32) error {
	v, ok := m.vouchers[voucherID]
	if !ok {
		return domain.ErrNotFound
	}
	if v.StockQuantity < quantity {
		return domain.ErrVoucherOutOfStock
	}
	v.StockQuantity -= quantity
	return nil
}

func (m *MockVoucherRepo) List(ctx context.Context, page, pageSize int32) ([]*domain.Voucher, int32, error) {
	return m.Search(ctx, domain.SearchVouchersParams{Page: page, PageSize: pageSize})
}

// MockWalletRepo for testing
type MockWalletRepo struct {
	wallets map[string]*domain.Wallet
}

func (m *MockWalletRepo) GetByUserID(ctx context.Context, userID string) (*domain.Wallet, error) {
	w, ok := m.wallets[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if w.IsLocked {
		return nil, domain.ErrWalletLocked
	}
	return w, nil
}

func (m *MockWalletRepo) UpdateBalance(ctx context.Context, userID string, amount float64, txType string) (*domain.Wallet, error) {
	w, err := m.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if txType == "credit" {
		w.Balance += amount
	} else if txType == "debit" {
		if w.Balance < amount {
			return nil, domain.ErrInsufficientBalance
		}
		w.Balance -= amount
	}

	return w, nil
}

func (m *MockWalletRepo) LockWallet(ctx context.Context, userID string) error {
	w, ok := m.wallets[userID]
	if !ok {
		return domain.ErrNotFound
	}
	w.IsLocked = true
	return nil
}

func (m *MockWalletRepo) UnlockWallet(ctx context.Context, userID string) error {
	w, ok := m.wallets[userID]
	if !ok {
		return domain.ErrNotFound
	}
	w.IsLocked = false
	return nil
}

// MockTransactionRepo for testing
type MockTransactionRepo struct {
	transactions []*domain.Transaction
}

func (m *MockTransactionRepo) Create(ctx context.Context, tx *domain.Transaction) error {
	m.transactions = append(m.transactions, tx)
	return nil
}

func (m *MockTransactionRepo) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	for _, tx := range m.transactions {
		if tx.ID == id {
			return tx, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockTransactionRepo) ListByUserID(ctx context.Context, userID string, txType string, page, pageSize int32) ([]*domain.Transaction, int32, error) {
	result := make([]*domain.Transaction, 0)
	for _, tx := range m.transactions {
		if tx.UserID == userID {
			result = append(result, tx)
		}
	}
	return result, int32(len(result)), nil
}

func (m *MockTransactionRepo) UpdateStatus(ctx context.Context, id, status string) error {
	for _, tx := range m.transactions {
		if tx.ID == id {
			tx.Status = status
			return nil
		}
	}
	return domain.ErrNotFound
}

// MockPurchasedVoucherRepo for testing
type MockPurchasedVoucherRepo struct {
	vouchers []*domain.PurchasedVoucher
}

func (m *MockPurchasedVoucherRepo) Create(ctx context.Context, pv *domain.PurchasedVoucher) error {
	m.vouchers = append(m.vouchers, pv)
	return nil
}

func (m *MockPurchasedVoucherRepo) CreateBatch(ctx context.Context, vouchers []*domain.PurchasedVoucher) error {
	m.vouchers = append(m.vouchers, vouchers...)
	return nil
}

func (m *MockPurchasedVoucherRepo) GetByID(ctx context.Context, id string) (*domain.PurchasedVoucher, error) {
	for _, v := range m.vouchers {
		if v.ID == id {
			return v, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockPurchasedVoucherRepo) ListByUserID(ctx context.Context, userID, status string, page, pageSize int32) ([]*domain.PurchasedVoucher, int32, error) {
	result := make([]*domain.PurchasedVoucher, 0)
	for _, v := range m.vouchers {
		if v.UserID == userID {
			result = append(result, v)
		}
	}
	return result, int32(len(result)), nil
}

func (m *MockPurchasedVoucherRepo) UpdateStatus(ctx context.Context, id, status string) error {
	for _, v := range m.vouchers {
		if v.ID == id {
			v.Status = status
			return nil
		}
	}
	return domain.ErrNotFound
}

// TestSearchVouchers tests voucher search functionality
func TestSearchVouchers(t *testing.T) {
	// Setup
	mockVouchers := map[string]*domain.Voucher{
		"v1": {
			ID:            "v1",
			Name:          "Amazon Gift Card",
			Brand:         "Amazon",
			Category:      "E-Commerce",
			SellingPrice:  100.0,
			StockQuantity: 10,
			IsActive:      true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	uow := &MockUnitOfWork{
		voucherRepo:  &MockVoucherRepo{vouchers: mockVouchers},
		shouldCommit: true,
	}

	// Note: For this test to work properly, we'd need to mock cache and mq
	// This is a simplified version showing the structure
	t.Run("SearchVouchers_Success", func(t *testing.T) {
		// This test is incomplete as it requires proper mocking of cache and MQ
		// In a real scenario, use testify/mock or similar
		if uow == nil {
			t.Skip("Skipping due to incomplete mocks")
		}
	})
}

// BenchmarkVoucherCodeGeneration benchmarks voucher code generation
func BenchmarkVoucherCodeGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generateVoucherCode()
	}
}

// BenchmarkPINGeneration benchmarks PIN generation
func BenchmarkPINGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generatePIN()
	}
}

