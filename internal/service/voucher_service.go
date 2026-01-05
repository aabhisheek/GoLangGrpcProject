package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/voucher-payment-service/internal/cache"
	"github.com/voucher-payment-service/internal/domain"
	"github.com/voucher-payment-service/internal/messaging"
	"github.com/voucher-payment-service/internal/repository"
	"go.uber.org/zap"
)

// VoucherService handles voucher business logic
type VoucherService struct {
	uow    repository.UnitOfWork
	cache  *cache.RedisCache
	mq     *messaging.RabbitMQ
	logger *zap.Logger
}

// NewVoucherService creates a new voucher service
func NewVoucherService(
	uow repository.UnitOfWork,
	cache *cache.RedisCache,
	mq *messaging.RabbitMQ,
	logger *zap.Logger,
) *VoucherService {
	return &VoucherService{
		uow:    uow,
		cache:  cache,
		mq:     mq,
		logger: logger,
	}
}

// SearchVouchers searches for vouchers
func (s *VoucherService) SearchVouchers(ctx context.Context, params domain.SearchVouchersParams) ([]*domain.Voucher, int32, error) {
	s.logger.Info("searching vouchers",
		zap.String("query", params.Query),
		zap.String("category", params.Category),
		zap.String("brand", params.Brand),
	)

	// Try cache first
	cacheKey := fmt.Sprintf("vouchers:search:%s:%s:%s:%f:%f:%d:%d",
		params.Query, params.Category, params.Brand,
		params.MinPrice, params.MaxPrice, params.Page, params.PageSize,
	)

	var cachedResult struct {
		Vouchers   []*domain.Voucher
		TotalCount int32
	}

	err := s.cache.Get(ctx, cacheKey, &cachedResult)
	if err == nil {
		s.logger.Debug("cache hit for voucher search", zap.String("key", cacheKey))
		return cachedResult.Vouchers, cachedResult.TotalCount, nil
	}

	// Cache miss, query database
	vouchers, totalCount, err := s.uow.Vouchers().Search(ctx, params)
	if err != nil {
		s.logger.Error("failed to search vouchers", zap.Error(err))
		return nil, 0, err
	}

	// Cache the result
	cachedResult.Vouchers = vouchers
	cachedResult.TotalCount = totalCount
	s.cache.Set(ctx, cacheKey, cachedResult, 5*time.Minute)

	return vouchers, totalCount, nil
}

// GetVoucher retrieves a voucher by ID
func (s *VoucherService) GetVoucher(ctx context.Context, voucherID string) (*domain.Voucher, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("voucher:%s", voucherID)
	
	var voucher domain.Voucher
	err := s.cache.Get(ctx, cacheKey, &voucher)
	if err == nil {
		return &voucher, nil
	}

	// Cache miss, query database
	voucherPtr, err := s.uow.Vouchers().GetByID(ctx, voucherID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.cache.Set(ctx, cacheKey, voucherPtr, 10*time.Minute)

	return voucherPtr, nil
}

// BuyVoucher handles voucher purchase
func (s *VoucherService) BuyVoucher(ctx context.Context, userID, voucherID string, quantity int32, paymentMethod, upiID string) (*domain.Transaction, []*domain.PurchasedVoucher, error) {
	s.logger.Info("buying voucher",
		zap.String("user_id", userID),
		zap.String("voucher_id", voucherID),
		zap.Int32("quantity", quantity),
		zap.String("payment_method", paymentMethod),
	)

	if quantity <= 0 {
		return nil, nil, domain.ErrInvalidInput
	}

	// Start transaction
	if err := s.uow.Begin(ctx); err != nil {
		return nil, nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			s.uow.Rollback()
			panic(r)
		}
	}()

	// Get voucher details
	voucher, err := s.uow.Vouchers().GetByID(ctx, voucherID)
	if err != nil {
		s.uow.Rollback()
		return nil, nil, err
	}

	if !voucher.IsActive {
		s.uow.Rollback()
		return nil, nil, domain.ErrVoucherInactive
	}

	if voucher.StockQuantity < quantity {
		s.uow.Rollback()
		return nil, nil, domain.ErrVoucherOutOfStock
	}

	totalAmount := voucher.SellingPrice * float64(quantity)

	// Handle payment
	var transaction *domain.Transaction
	if paymentMethod == "wallet" {
		// Deduct from wallet
		wallet, err := s.uow.Wallets().GetByUserID(ctx, userID)
		if err != nil {
			s.uow.Rollback()
			return nil, nil, err
		}

		if wallet.Balance < totalAmount {
			s.uow.Rollback()
			return nil, nil, domain.ErrInsufficientBalance
		}

		updatedWallet, err := s.uow.Wallets().UpdateBalance(ctx, userID, totalAmount, "debit")
		if err != nil {
			s.uow.Rollback()
			return nil, nil, err
		}

		// Create transaction record
		transaction = &domain.Transaction{
			ID:            generateID("txn"),
			UserID:        userID,
			Type:          "debit",
			Amount:        totalAmount,
			BalanceBefore: wallet.Balance,
			BalanceAfter:  updatedWallet.Balance,
			Description:   fmt.Sprintf("Voucher purchase: %s x%d", voucher.Name, quantity),
			ReferenceID:   voucherID,
			PaymentMethod: "wallet",
			Status:        "completed",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
	} else if paymentMethod == "upi" {
		// Mock UPI payment
		transaction = &domain.Transaction{
			ID:            generateID("txn"),
			UserID:        userID,
			Type:          "debit",
			Amount:        totalAmount,
			BalanceBefore: 0,
			BalanceAfter:  0,
			Description:   fmt.Sprintf("Voucher purchase via UPI: %s x%d", voucher.Name, quantity),
			ReferenceID:   voucherID,
			PaymentMethod: fmt.Sprintf("upi:%s", upiID),
			Status:        "completed",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
	} else {
		s.uow.Rollback()
		return nil, nil, domain.ErrInvalidPaymentMethod
	}

	err = s.uow.Transactions().Create(ctx, transaction)
	if err != nil {
		s.uow.Rollback()
		return nil, nil, err
	}

	// Update stock
	err = s.uow.Vouchers().UpdateStock(ctx, voucherID, quantity)
	if err != nil {
		s.uow.Rollback()
		return nil, nil, err
	}

	// Create purchased vouchers
	purchasedVouchers := make([]*domain.PurchasedVoucher, quantity)
	voucherCodes := make([]string, quantity)
	
	for i := int32(0); i < quantity; i++ {
		code := generateVoucherCode()
		pin := generatePIN()
		
		pv := &domain.PurchasedVoucher{
			ID:            generateID("pv"),
			UserID:        userID,
			VoucherID:     voucherID,
			TransactionID: transaction.ID,
			VoucherCode:   code,
			Pin:           pin,
			Status:        "active",
			PurchasePrice: voucher.SellingPrice,
			ValidUntil:    voucher.ValidUntil,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		purchasedVouchers[i] = pv
		voucherCodes[i] = code
	}

	err = s.uow.PurchasedVouchers().CreateBatch(ctx, purchasedVouchers)
	if err != nil {
		s.uow.Rollback()
		return nil, nil, err
	}

	// Commit transaction
	if err := s.uow.Commit(); err != nil {
		s.uow.Rollback()
		return nil, nil, err
	}

	// Invalidate cache
	s.cache.DeletePattern(ctx, "vouchers:*")
	s.cache.Delete(ctx, fmt.Sprintf("voucher:%s", voucherID))
	s.cache.Delete(ctx, fmt.Sprintf("wallet:%s", userID))

	// Publish event to message queue
	event := domain.VoucherPurchaseEvent{
		TransactionID: transaction.ID,
		UserID:        userID,
		VoucherID:     voucherID,
		Quantity:      quantity,
		TotalAmount:   totalAmount,
		VoucherCodes:  voucherCodes,
		Timestamp:     time.Now().Unix(),
	}

	go func() {
		ctx := context.Background()
		err := s.mq.Publish(ctx, s.mq.GetQueueName(), event)
		if err != nil {
			s.logger.Error("failed to publish purchase event", zap.Error(err))
		}
	}()

	s.logger.Info("voucher purchased successfully",
		zap.String("transaction_id", transaction.ID),
		zap.Int32("quantity", quantity),
		zap.Float64("amount", totalAmount),
	)

	return transaction, purchasedVouchers, nil
}

// ListUserVouchers lists vouchers owned by a user
func (s *VoucherService) ListUserVouchers(ctx context.Context, userID, status string, page, pageSize int32) ([]*domain.PurchasedVoucher, int32, error) {
	return s.uow.PurchasedVouchers().ListByUserID(ctx, userID, status, page, pageSize)
}

// Helper functions

func generateID(prefix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d", prefix, timestamp)
}

func generateVoucherCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 16)
	
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code[i] = chars[n.Int64()]
		
		// Add hyphen every 4 characters
		if (i+1)%4 == 0 && i < 15 {
			code = append(code[:i+1], append([]byte{'-'}, code[i+1:]...)...)
			i++
		}
	}
	
	return string(code[:19]) // 16 chars + 3 hyphens
}

func generatePIN() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("%04d", n.Int64())
}

