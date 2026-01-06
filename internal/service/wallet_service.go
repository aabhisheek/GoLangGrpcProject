package service

import (
	"context"
	"fmt"
	"time"

	"github.com/voucher-payment-service/internal/cache"
	"github.com/voucher-payment-service/internal/domain"
	"github.com/voucher-payment-service/internal/repository"
	"go.uber.org/zap"
)

// WalletService handles wallet business logic
type WalletService struct {
	uow    repository.UnitOfWork
	cache  *cache.RedisCache
	logger *zap.Logger
}

// NewWalletService creates a new wallet service
func NewWalletService(
	uow repository.UnitOfWork,
	cache *cache.RedisCache,
	logger *zap.Logger,
) *WalletService {
	return &WalletService{
		uow:    uow,
		cache:  cache,
		logger: logger,
	}
}

// GetBalance retrieves wallet balance for a user
func (s *WalletService) GetBalance(ctx context.Context, userID string) (*domain.Wallet, error) {
	s.logger.Info("getting wallet balance", zap.String("user_id", userID))

	// Try cache first
	cacheKey := fmt.Sprintf("wallet:%s", userID)
	
	var wallet domain.Wallet
	err := s.cache.Get(ctx, cacheKey, &wallet)
	if err == nil {
		return &wallet, nil
	}

	// Cache miss, query database (using read-only method to avoid transaction requirement)
	walletPtr, err := s.uow.Wallets().GetByUserIDReadOnly(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get wallet", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}

	// Cache the result
	s.cache.Set(ctx, cacheKey, walletPtr, 5*time.Minute)

	return walletPtr, nil
}

// AddMoney adds money to a user's wallet
func (s *WalletService) AddMoney(ctx context.Context, userID string, amount float64, paymentMethod, transactionRef string) (*domain.Transaction, *domain.Wallet, error) {
	s.logger.Info("adding money to wallet",
		zap.String("user_id", userID),
		zap.Float64("amount", amount),
		zap.String("payment_method", paymentMethod),
	)

	if amount <= 0 {
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

	// Get current wallet
	wallet, err := s.uow.Wallets().GetByUserID(ctx, userID)
	if err != nil {
		s.uow.Rollback()
		return nil, nil, err
	}

	// Update balance
	updatedWallet, err := s.uow.Wallets().UpdateBalance(ctx, userID, amount, "credit")
	if err != nil {
		s.uow.Rollback()
		return nil, nil, err
	}

	// Create transaction record
	transaction := &domain.Transaction{
		ID:            generateID("txn"),
		UserID:        userID,
		Type:          "credit",
		Amount:        amount,
		BalanceBefore: wallet.Balance,
		BalanceAfter:  updatedWallet.Balance,
		Description:   fmt.Sprintf("Money added via %s", paymentMethod),
		ReferenceID:   transactionRef,
		PaymentMethod: paymentMethod,
		Status:        "completed",
		Metadata:      "{}",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = s.uow.Transactions().Create(ctx, transaction)
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
	s.cache.Delete(ctx, fmt.Sprintf("wallet:%s", userID))

	s.logger.Info("money added successfully",
		zap.String("transaction_id", transaction.ID),
		zap.Float64("amount", amount),
		zap.Float64("new_balance", updatedWallet.Balance),
	)

	return transaction, updatedWallet, nil
}

// ListTransactions lists transactions for a user
func (s *WalletService) ListTransactions(ctx context.Context, userID, txType string, page, pageSize int32) ([]*domain.Transaction, int32, error) {
	s.logger.Info("listing transactions",
		zap.String("user_id", userID),
		zap.String("type", txType),
		zap.Int32("page", page),
		zap.Int32("page_size", pageSize),
	)

	// Try cache first for recent transactions
	if page == 1 && txType == "all" {
		cacheKey := fmt.Sprintf("transactions:%s:recent", userID)
		
		var cachedResult struct {
			Transactions []*domain.Transaction
			TotalCount   int32
		}

		err := s.cache.Get(ctx, cacheKey, &cachedResult)
		if err == nil {
			return cachedResult.Transactions, cachedResult.TotalCount, nil
		}
	}

	transactions, totalCount, err := s.uow.Transactions().ListByUserID(ctx, userID, txType, page, pageSize)
	if err != nil {
		s.logger.Error("failed to list transactions", zap.String("user_id", userID), zap.Error(err))
		return nil, 0, err
	}

	// Cache recent transactions
	if page == 1 && txType == "all" {
		cacheKey := fmt.Sprintf("transactions:%s:recent", userID)
		cachedResult := struct {
			Transactions []*domain.Transaction
			TotalCount   int32
		}{
			Transactions: transactions,
			TotalCount:   totalCount,
		}
		s.cache.Set(ctx, cacheKey, cachedResult, 2*time.Minute)
	}

	return transactions, totalCount, nil
}

// GetTransaction retrieves a transaction by ID
func (s *WalletService) GetTransaction(ctx context.Context, transactionID string) (*domain.Transaction, error) {
	return s.uow.Transactions().GetByID(ctx, transactionID)
}

