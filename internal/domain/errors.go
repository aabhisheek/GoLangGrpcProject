package domain

import (
	"errors"
	"fmt"
)

// Common domain errors
var (
	ErrNotFound              = errors.New("resource not found")
	ErrInvalidInput          = errors.New("invalid input")
	ErrInsufficientBalance   = errors.New("insufficient wallet balance")
	ErrVoucherOutOfStock     = errors.New("voucher out of stock")
	ErrVoucherInactive       = errors.New("voucher is not active")
	ErrVoucherExpired        = errors.New("voucher has expired")
	ErrWalletLocked          = errors.New("wallet is locked")
	ErrTransactionFailed     = errors.New("transaction failed")
	ErrDuplicateTransaction  = errors.New("duplicate transaction")
	ErrInvalidPaymentMethod  = errors.New("invalid payment method")
	ErrUnauthorized          = errors.New("unauthorized access")
)

// AppError represents a custom application error
type AppError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error codes
const (
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeInvalidInput        = "INVALID_INPUT"
	ErrCodeInsufficientBalance = "INSUFFICIENT_BALANCE"
	ErrCodeOutOfStock          = "OUT_OF_STOCK"
	ErrCodeInactive            = "INACTIVE"
	ErrCodeExpired             = "EXPIRED"
	ErrCodeLocked              = "LOCKED"
	ErrCodeTransactionFailed   = "TRANSACTION_FAILED"
	ErrCodeDuplicate           = "DUPLICATE"
	ErrCodeInvalidPayment      = "INVALID_PAYMENT"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeInternal            = "INTERNAL_ERROR"
)

