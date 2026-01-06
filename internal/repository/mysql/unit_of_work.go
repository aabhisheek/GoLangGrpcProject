package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voucher-payment-service/internal/repository"
)

// UnitOfWork implements repository.UnitOfWork for MySQL
type UnitOfWork struct {
	db                         *DB
	tx                         *sql.Tx
	voucherRepo                *VoucherRepository
	walletRepo                 *WalletRepository
	transactionRepo            *TransactionRepository
	purchasedVoucherRepo       *PurchasedVoucherRepository
	userRepo                   *UserRepository
	paymentEventRepo           *PaymentEventRepository
}

// NewUnitOfWork creates a new unit of work
func NewUnitOfWork(db *DB) repository.UnitOfWork {
	return &UnitOfWork{
		db: db,
	}
}

// Begin starts a new transaction
func (uow *UnitOfWork) Begin(ctx context.Context) error {
	tx, err := uow.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	uow.tx = tx

	// Initialize repositories with transaction
	uow.voucherRepo = NewVoucherRepository(uow.db.DB).WithTx(tx)
	uow.walletRepo = NewWalletRepository(uow.db.DB).WithTx(tx)
	uow.transactionRepo = NewTransactionRepository(uow.db.DB).WithTx(tx)
	uow.purchasedVoucherRepo = NewPurchasedVoucherRepository(uow.db.DB).WithTx(tx)
	uow.userRepo = NewUserRepository(uow.db.DB).WithTx(tx)
	uow.paymentEventRepo = NewPaymentEventRepository(uow.db.DB).WithTx(tx)

	return nil
}

// Commit commits the transaction
func (uow *UnitOfWork) Commit() error {
	if uow.tx == nil {
		return fmt.Errorf("no active transaction")
	}

	err := uow.tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	uow.tx = nil
	// Clear all repository instances so they get recreated without transaction
	uow.voucherRepo = nil
	uow.walletRepo = nil
	uow.transactionRepo = nil
	uow.purchasedVoucherRepo = nil
	uow.userRepo = nil
	uow.paymentEventRepo = nil
	
	return nil
}

// Rollback rolls back the transaction
func (uow *UnitOfWork) Rollback() error {
	if uow.tx == nil {
		return fmt.Errorf("no active transaction")
	}

	err := uow.tx.Rollback()
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	uow.tx = nil
	// Clear all repository instances so they get recreated without transaction
	uow.voucherRepo = nil
	uow.walletRepo = nil
	uow.transactionRepo = nil
	uow.purchasedVoucherRepo = nil
	uow.userRepo = nil
	uow.paymentEventRepo = nil
	
	return nil
}

// Vouchers returns the voucher repository
func (uow *UnitOfWork) Vouchers() repository.VoucherRepository {
	if uow.voucherRepo == nil {
		uow.voucherRepo = NewVoucherRepository(uow.db.DB)
	}
	return uow.voucherRepo
}

// Wallets returns the wallet repository
func (uow *UnitOfWork) Wallets() repository.WalletRepository {
	if uow.walletRepo == nil {
		uow.walletRepo = NewWalletRepository(uow.db.DB)
	}
	return uow.walletRepo
}

// Transactions returns the transaction repository
func (uow *UnitOfWork) Transactions() repository.TransactionRepository {
	if uow.transactionRepo == nil {
		uow.transactionRepo = NewTransactionRepository(uow.db.DB)
	}
	return uow.transactionRepo
}

// PurchasedVouchers returns the purchased voucher repository
func (uow *UnitOfWork) PurchasedVouchers() repository.PurchasedVoucherRepository {
	if uow.purchasedVoucherRepo == nil {
		uow.purchasedVoucherRepo = NewPurchasedVoucherRepository(uow.db.DB)
	}
	return uow.purchasedVoucherRepo
}

// Users returns the user repository
func (uow *UnitOfWork) Users() repository.UserRepository {
	if uow.userRepo == nil {
		uow.userRepo = NewUserRepository(uow.db.DB)
	}
	return uow.userRepo
}

// PaymentEvents returns the payment event repository
func (uow *UnitOfWork) PaymentEvents() repository.PaymentEventRepository {
	if uow.paymentEventRepo == nil {
		uow.paymentEventRepo = NewPaymentEventRepository(uow.db.DB)
	}
	return uow.paymentEventRepo
}

