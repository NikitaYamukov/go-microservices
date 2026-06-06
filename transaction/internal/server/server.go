package server

import (
	"context"

	paginationpb "github.com/NikitaYamukov/contracts/pagination/go"
	transactionpb "github.com/NikitaYamukov/contracts/transaction/go"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"
	"transaction/internal/model"
)

type Server struct {
	transactionpb.UnimplementedTransactionServiceServer

	transactionService TransactionService
	logger             *zerolog.Logger
}

func New(transactionService TransactionService, logger *zerolog.Logger) *Server {
	return &Server{transactionService: transactionService, logger: logger}
}

type TransactionService interface {
	GetTransactionsWithDetails(context.Context, model.GetTransactionsParams) ([]model.TransactionDetails, error)
	Deposit(context.Context, uint64, float64) (model.TransactionDetails, error)
	Withdraw(context.Context, uint64, float64) (model.TransactionDetails, error)
	Transfer(context.Context, uint64, uint64, float64) (model.TransactionDetails, error)
}

func (s *Server) Deposit(ctx context.Context, req *transactionpb.DepositRequest) (*transactionpb.TransactionDetails, error) {
	res, err := s.transactionService.Deposit(ctx, req.GetUserId(), req.GetAmount())
	if err != nil {
		return nil, err
	}

	return TransactionDetailsToPb(res), nil
}

func (s *Server) Withdraw(ctx context.Context, req *transactionpb.WithdrawRequest) (*transactionpb.TransactionDetails, error) {
	res, err := s.transactionService.Withdraw(ctx, req.GetAccountId(), req.GetAmount())
	if err != nil {
		return nil, err
	}

	return TransactionDetailsToPb(res), nil
}

func (s *Server) Transfer(ctx context.Context, req *transactionpb.TransferRequest) (*transactionpb.TransactionDetails, error) {
	res, err := s.transactionService.Transfer(ctx, req.GetUserId(), req.GetRecipient(), req.GetAmount())
	if err != nil {
		return nil, err
	}

	return TransactionDetailsToPb(res), nil
}

func (s *Server) GetTransactions(ctx context.Context, req *transactionpb.GetTransactionsRequest) (
	*transactionpb.GetTransactionsResponse, error) {
	params := model.GetTransactionsParams{}
	if req.UserId != nil {
		params.UserID = req.UserId
	}
	if req.Type != nil {
		params.Type = req.Type
	}
	if req.Status != nil {
		params.Status = req.Status
	}
	if req.DateFrom != nil {
		dateFrom := req.DateFrom.AsTime()
		params.DateFrom = &dateFrom
	}
	if req.DateTo != nil {
		dateTo := req.DateTo.AsTime()
		params.DateTo = &dateTo
	}
	if req.GetPagination() != nil {
		params.Limit = int(req.GetPagination().GetLimit())
		params.Offset = int(req.GetPagination().GetOffset())
	}

	res, err := s.transactionService.GetTransactionsWithDetails(ctx, params)
	if err != nil {
		return nil, err
	}

	return &transactionpb.GetTransactionsResponse{
		Transactions: TransactionDetailsListToPb(res),
		Pagination:   req.GetPagination(),
	}, nil
}

func TransactionDetailsListToPb(details []model.TransactionDetails) []*transactionpb.TransactionDetails {
	res := make([]*transactionpb.TransactionDetails, len(details))
	for i, detail := range details {
		res[i] = TransactionDetailsToPb(detail)
	}

	return res
}

func TransactionDetailsToPb(details model.TransactionDetails) *transactionpb.TransactionDetails {
	return &transactionpb.TransactionDetails{
		Transaction: TransactionToPb(details.Transaction),
		Entries:     TransactionEntriesToPb(details.Entries),
	}
}

func TransactionToPb(transaction model.Transaction) *transactionpb.Transaction {
	return &transactionpb.Transaction{
		Id:        transaction.ID,
		UserId:    transaction.UserID,
		Amount:    float64(transaction.Amount),
		Status:    string(transaction.Status),
		Type:      string(transaction.Type),
		CreatedAt: timestamppb.New(transaction.CreatedAt),
		UpdatedAt: timestamppb.New(transaction.UpdatedAt),
	}
}

func TransactionEntriesToPb(entries []model.TransactionEntry) []*transactionpb.TransactionEntry {
	res := make([]*transactionpb.TransactionEntry, len(entries))
	for i, entry := range entries {
		res[i] = TransactionEntryToPb(entry)
	}

	return res
}

func TransactionEntryToPb(entry model.TransactionEntry) *transactionpb.TransactionEntry {
	return &transactionpb.TransactionEntry{
		Id:            entry.ID,
		TransactionId: entry.TransactionID,
		AccountId:     entry.AccountID,
		Direction:     string(entry.Direction),
		Amount:        entry.Amount,
		CreatedAt:     timestamppb.New(entry.CreatedAt),
		UpdatedAt:     timestamppb.New(entry.UpdatedAt),
	}
}

func PaginationToPb(limit, offset uint32) *paginationpb.Pagination {
	return &paginationpb.Pagination{
		Limit:  limit,
		Offset: offset,
	}
}
