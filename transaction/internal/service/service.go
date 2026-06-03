package service

import (
	"context"

	"github.com/NikitaYamukov/go-microservices/transaction/internal/model"
	"github.com/rs/zerolog"
)

type TransactionService struct {
	repo   Repository
	logger *zerolog.Logger
}

func New(repo Repository, logger *zerolog.Logger) *TransactionService {
	return &TransactionService{
		repo:   repo,
		logger: logger,
	}
}

type Repository interface {
	// Transaction methods
	GetTransactions(context.Context, model.GetTransactionsParams) ([]model.Transaction, error)

	// TransactionEntry methods
	GetTransactionDetails(context.Context, uint64) (model.TransactionDetails, error)

	// Business operations
	Deposit(context.Context, model.DepositParams) (model.TransactionDetails, error)
	Withdraw(context.Context, model.WithdrawParams) (model.TransactionDetails, error)
	Transfer(context.Context, model.TransferParams) (model.TransactionDetails, error)
}

func (s *TransactionService) GetTransactionsWithDetails(
	ctx context.Context,
	params model.GetTransactionsParams,
) ([]model.TransactionDetails, error) {
	// Получаем список транзакций с фильтрацией
	transactions, err := s.repo.GetTransactions(ctx, params)
	if err != nil {
		s.logger.Err(err).Msg("failed to get transactions")
		return nil, err
	}

	// Для каждой транзакции получаем детали
	transactionDetails := []model.TransactionDetails{}
	for _, transaction := range transactions {
		details, err := s.repo.GetTransactionDetails(ctx, transaction.ID)
		if err != nil {
			s.logger.Err(err).Uint64("transaction_id", transaction.ID).Msg("failed to get transaction details")
			continue
		}
		transactionDetails = append(transactionDetails, details)
	}

	return transactionDetails, nil
}

// Deposit - Business logic methods
func (s *TransactionService) Deposit(
	ctx context.Context,
	userID uint64,
	amount float64,
) (model.TransactionDetails, error) {
	params := model.DepositParams{
		UserID: userID,
		Amount: amount,
	}

	return s.repo.Deposit(ctx, params)
}

func (s *TransactionService) Withdraw(
	ctx context.Context,
	accountID uint64,
	amount float64,
) (model.TransactionDetails, error) {
	params := model.WithdrawParams{
		AccountID: accountID,
		Amount:    amount,
	}

	return s.repo.Withdraw(ctx, params)
}

func (s *TransactionService) Transfer(
	ctx context.Context,
	userID uint64,
	recipient uint64,
	amount float64,
) (model.TransactionDetails, error) {
	params := model.TransferParams{
		UserID:    userID,
		Recipient: recipient,
		Amount:    amount,
	}

	return s.repo.Transfer(ctx, params)
}
