package service

import (
	"context"

	"github.com/NikitaYamukov/go-microservices/transaction/internal/model"
	"github.com/rs/zerolog"
)

type TransactionService struct {
	repo           Repository
	logger         *zerolog.Logger
	accountService AccountService
}

func New(repo Repository, accountService AccountService, logger *zerolog.Logger) *TransactionService {
	return &TransactionService{
		repo:           repo,
		accountService: accountService,
		logger:         logger,
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

	// Update transaction status
	UpdateTransactionStatus(context.Context, uint64, model.TransactionStatus) error
}

type AccountService interface {
	GetBalance(ctx context.Context, userID uint64) (int64, error)
	Deposit(ctx context.Context, userID uint64, amount int64, operationID uint64) (int64, error)
	Withdraw(ctx context.Context, userID uint64, amount int64, operationID uint64) (int64, error)
	Transfer(ctx context.Context, fromAccountID, toAccountID uint64, amount int64, operationID uint64) (string, error)
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

	// Создаем транзакцию в pending статусе
	res, err := s.repo.Deposit(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Uint64("user_id", userID).Msg("failed to create deposit transaction")
		return model.TransactionDetails{}, err
	}

	// Выполняем операцию с балансом через account service
	_, err = s.accountService.Deposit(ctx, userID, int64(amount*100), res.Transaction.ID)
	if err != nil {
		s.logger.Error().Err(err).Uint64("user_id", userID).Uint64("transaction_id",
			res.Transaction.ID).Msg("failed to deposit to account")

		// Обновляем статус транзакции на failed
		updateErr := s.repo.UpdateTransactionStatus(ctx, res.Transaction.ID, model.TransactionStatusFailed)

		if updateErr != nil {
			s.logger.Error().Err(updateErr).Uint64("transaction_id",
				res.Transaction.ID).Msg("failed to update transaction status to failed")
		}

		return model.TransactionDetails{}, err
	}

	// Обновляем статус транзакции на completed
	err = s.repo.UpdateTransactionStatus(ctx, res.Transaction.ID, model.TransactionStatusCompleted)
	if err != nil {
		s.logger.Error().Err(err).Uint64("transaction_id", res.Transaction.ID).Msg("failed to update transaction status to completed")
		return model.TransactionDetails{}, err
	}

	// Обновляем статус в возвращаемом результате
	res.Transaction.Status = model.TransactionStatusCompleted
	return res, nil
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

	// Создаем транзакцию в pending статусе
	res, err := s.repo.Withdraw(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Uint64("account_id", accountID).Msg("failed to create withdraw transaction")
		return model.TransactionDetails{}, err
	}

	// Выполняем операцию с балансом через account service
	_, err = s.accountService.Withdraw(ctx, accountID, int64(amount*100), res.Transaction.ID)
	if err != nil {
		s.logger.Error().Err(err).Uint64("account_id", accountID).Uint64("transaction_id",
			res.Transaction.ID).Msg("failed to withdraw from account")

		// Обновляем статус транзакции на failed
		updateErr := s.repo.UpdateTransactionStatus(ctx, res.Transaction.ID,
			model.TransactionStatusFailed)

		if updateErr != nil {
			s.logger.Error().Err(updateErr).Uint64("transaction_id",
				res.Transaction.ID).Msg("failed to update transaction status to failed")
		}

		return model.TransactionDetails{}, err
	}

	// Обновляем статус транзакции на completed
	err = s.repo.UpdateTransactionStatus(ctx, res.Transaction.ID,
		model.TransactionStatusCompleted)
	if err != nil {
		s.logger.Error().Err(err).Uint64("transaction_id", res.Transaction.ID).Msg("failed to update transaction status to completed")
		return model.TransactionDetails{}, err
	}

	// Обновляем статус в возвращаемом результате
	res.Transaction.Status = model.TransactionStatusCompleted
	return res, nil
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

	// Создаем транзакцию в pending статусе
	res, err := s.repo.Transfer(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Uint64("user_id", userID).Uint64("recipient",
			recipient).Msg("failed to create transfer transaction")
		return model.TransactionDetails{}, err
	}

	// Выполняем операцию с балансом через account service
	_, err = s.accountService.Transfer(ctx, userID, recipient, int64(amount*100), res.Transaction.ID)
	if err != nil {
		s.logger.Error().Err(err).Uint64("user_id", userID).Uint64("recipient",
			recipient).Uint64("transaction_id", res.Transaction.ID).Msg("failed to transfer between accounts")

		// Обновляем статус транзакции на failed
		updateErr := s.repo.UpdateTransactionStatus(ctx, res.Transaction.ID,
			model.TransactionStatusFailed)

		if updateErr != nil {
			s.logger.Error().Err(updateErr).Uint64("transaction_id",
				res.Transaction.ID).Msg("failed to update transaction status to failed")
		}

		return model.TransactionDetails{}, err
	}

	// Обновляем статус транзакции на completed
	err = s.repo.UpdateTransactionStatus(ctx, res.Transaction.ID,
		model.TransactionStatusCompleted)
	if err != nil {
		s.logger.Error().Err(err).Uint64("transaction_id", res.Transaction.ID).Msg("failed to update transaction status to completed")
		return model.TransactionDetails{}, err
	}

	// Обновляем статус в возвращаемом результате
	res.Transaction.Status = model.TransactionStatusCompleted
	return res, nil
}
