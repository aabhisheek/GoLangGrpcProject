package grpc

import (
	"context"

	voucherv1 "github.com/voucher-payment-service/api/generated"
	"github.com/voucher-payment-service/internal/domain"
	"github.com/voucher-payment-service/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// VoucherServer implements the gRPC VoucherService
type VoucherServer struct {
	voucherv1.UnimplementedVoucherServiceServer
	voucherService *service.VoucherService
	logger         *zap.Logger
}

// NewVoucherServer creates a new voucher server
func NewVoucherServer(voucherService *service.VoucherService, logger *zap.Logger) *VoucherServer {
	return &VoucherServer{
		voucherService: voucherService,
		logger:         logger,
	}
}

// SearchVouchers handles voucher search
func (s *VoucherServer) SearchVouchers(ctx context.Context, req *voucherv1.SearchVouchersRequest) (*voucherv1.SearchVouchersResponse, error) {
	params := domain.SearchVouchersParams{
		Query:    req.Query,
		Category: req.Category,
		Brand:    req.Brand,
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	vouchers, totalCount, err := s.voucherService.SearchVouchers(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search vouchers: %v", err)
	}

	protoVouchers := make([]*voucherv1.Voucher, len(vouchers))
	for i, v := range vouchers {
		protoVouchers[i] = toProtoVoucher(v)
	}

	return &voucherv1.SearchVouchersResponse{
		Vouchers:   protoVouchers,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// GetVoucher retrieves a single voucher
func (s *VoucherServer) GetVoucher(ctx context.Context, req *voucherv1.GetVoucherRequest) (*voucherv1.Voucher, error) {
	voucher, err := s.voucherService.GetVoucher(ctx, req.VoucherId)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "voucher not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get voucher: %v", err)
	}

	return toProtoVoucher(voucher), nil
}

// BuyVoucher handles voucher purchase
func (s *VoucherServer) BuyVoucher(ctx context.Context, req *voucherv1.BuyVoucherRequest) (*voucherv1.BuyVoucherResponse, error) {
	transaction, purchasedVouchers, err := s.voucherService.BuyVoucher(
		ctx,
		req.UserId,
		req.VoucherId,
		req.Quantity,
		req.PaymentMethod,
		req.UpiId,
	)
	if err != nil {
		switch err {
		case domain.ErrInsufficientBalance:
			return nil, status.Errorf(codes.FailedPrecondition, "insufficient balance")
		case domain.ErrVoucherOutOfStock:
			return nil, status.Errorf(codes.FailedPrecondition, "voucher out of stock")
		case domain.ErrVoucherInactive:
			return nil, status.Errorf(codes.FailedPrecondition, "voucher is not active")
		case domain.ErrInvalidPaymentMethod:
			return nil, status.Errorf(codes.InvalidArgument, "invalid payment method")
		default:
			return nil, status.Errorf(codes.Internal, "failed to buy voucher: %v", err)
		}
	}

	protoPurchasedVouchers := make([]*voucherv1.PurchasedVoucher, len(purchasedVouchers))
	for i, pv := range purchasedVouchers {
		protoPurchasedVouchers[i] = toProtoPurchasedVoucher(pv)
	}

	return &voucherv1.BuyVoucherResponse{
		TransactionId:      transaction.ID,
		Status:             transaction.Status,
		AmountPaid:         transaction.Amount,
		PurchasedVouchers:  protoPurchasedVouchers,
		PurchasedAt:        timestamppb.New(transaction.CreatedAt),
	}, nil
}

// ListUserVouchers lists user's vouchers
func (s *VoucherServer) ListUserVouchers(ctx context.Context, req *voucherv1.ListUserVouchersRequest) (*voucherv1.ListUserVouchersResponse, error) {
	vouchers, totalCount, err := s.voucherService.ListUserVouchers(ctx, req.UserId, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list user vouchers: %v", err)
	}

	protoVouchers := make([]*voucherv1.PurchasedVoucher, len(vouchers))
	for i, v := range vouchers {
		protoVouchers[i] = toProtoPurchasedVoucher(v)
	}

	return &voucherv1.ListUserVouchersResponse{
		Vouchers:   protoVouchers,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// WalletServer implements the gRPC WalletService
type WalletServer struct {
	voucherv1.UnimplementedWalletServiceServer
	walletService *service.WalletService
	logger        *zap.Logger
}

// NewWalletServer creates a new wallet server
func NewWalletServer(walletService *service.WalletService, logger *zap.Logger) *WalletServer {
	return &WalletServer{
		walletService: walletService,
		logger:        logger,
	}
}

// GetBalance retrieves wallet balance
func (s *WalletServer) GetBalance(ctx context.Context, req *voucherv1.GetBalanceRequest) (*voucherv1.GetBalanceResponse, error) {
	wallet, err := s.walletService.GetBalance(ctx, req.UserId)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "wallet not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get balance: %v", err)
	}

	return &voucherv1.GetBalanceResponse{
		UserId:   wallet.UserID,
		Balance:  wallet.Balance,
		Currency: wallet.Currency,
	}, nil
}

// AddMoney adds money to wallet
func (s *WalletServer) AddMoney(ctx context.Context, req *voucherv1.AddMoneyRequest) (*voucherv1.AddMoneyResponse, error) {
	transaction, wallet, err := s.walletService.AddMoney(
		ctx,
		req.UserId,
		req.Amount,
		req.PaymentMethod,
		req.TransactionReference,
	)
	if err != nil {
		switch err {
		case domain.ErrInvalidInput:
			return nil, status.Errorf(codes.InvalidArgument, "invalid amount")
		default:
			return nil, status.Errorf(codes.Internal, "failed to add money: %v", err)
		}
	}

	return &voucherv1.AddMoneyResponse{
		TransactionId: transaction.ID,
		Status:        transaction.Status,
		NewBalance:    wallet.Balance,
		CompletedAt:   timestamppb.New(transaction.CreatedAt),
	}, nil
}

// ListTransactions lists wallet transactions
func (s *WalletServer) ListTransactions(ctx context.Context, req *voucherv1.ListTransactionsRequest) (*voucherv1.ListTransactionsResponse, error) {
	transactions, totalCount, err := s.walletService.ListTransactions(ctx, req.UserId, req.Type, req.Page, req.PageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list transactions: %v", err)
	}

	protoTransactions := make([]*voucherv1.Transaction, len(transactions))
	for i, t := range transactions {
		protoTransactions[i] = toProtoTransaction(t)
	}

	return &voucherv1.ListTransactionsResponse{
		Transactions: protoTransactions,
		TotalCount:   totalCount,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}, nil
}

// Helper functions to convert domain models to proto messages

func toProtoVoucher(v *domain.Voucher) *voucherv1.Voucher {
	return &voucherv1.Voucher{
		Id:                 v.ID,
		Name:               v.Name,
		Description:        v.Description,
		Brand:              v.Brand,
		Category:           v.Category,
		FaceValue:          v.FaceValue,
		DiscountPercentage: v.DiscountPercentage,
		SellingPrice:       v.SellingPrice,
		StockQuantity:      v.StockQuantity,
		IsActive:           v.IsActive,
		ValidFrom:          timestamppb.New(v.ValidFrom),
		ValidUntil:         timestamppb.New(v.ValidUntil),
		TermsAndConditions: v.TermsAndConditions,
		CreatedAt:          timestamppb.New(v.CreatedAt),
		UpdatedAt:          timestamppb.New(v.UpdatedAt),
	}
}

func toProtoPurchasedVoucher(pv *domain.PurchasedVoucher) *voucherv1.PurchasedVoucher {
	return &voucherv1.PurchasedVoucher{
		Id:          pv.ID,
		VoucherId:   pv.VoucherID,
		VoucherCode: pv.VoucherCode,
		Pin:         pv.Pin,
		Status:      pv.Status,
		ValidUntil:  timestamppb.New(pv.ValidUntil),
	}
}

func toProtoTransaction(t *domain.Transaction) *voucherv1.Transaction {
	return &voucherv1.Transaction{
		Id:            t.ID,
		UserId:        t.UserID,
		Type:          t.Type,
		Amount:        t.Amount,
		BalanceBefore: t.BalanceBefore,
		BalanceAfter:  t.BalanceAfter,
		Description:   t.Description,
		ReferenceId:   t.ReferenceID,
		Status:        t.Status,
		CreatedAt:     timestamppb.New(t.CreatedAt),
	}
}

