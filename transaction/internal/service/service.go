package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"transaction/internal/model"
)

type TransactionService struct {
	repo           Repository
	logger         *zerolog.Logger
	accountService AccountService

	// Kafka support
	kafkaPublisher KafkaPublisher
}

func New(repo Repository, accountService AccountService, kafkaPublisher KafkaPublisher,
	logger *zerolog.Logger) *TransactionService {
	return &TransactionService{
		repo:           repo,
		accountService: accountService,
		kafkaPublisher: kafkaPublisher,
		logger:         logger,
	}
}

// AccountResponse представляет структуру ответа от account service
type AccountResponse struct {
	RequestType string `json:"request_type"`
	UserID      uint64 `json:"user_id"`
	OperationID uint64 `json:"operation_id"`
	Result      bool   `json:"result"`
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

// KafkaPublisher интерфейс для публикации сообщений в Kafka
type KafkaPublisher interface {
	Publish(ctx context.Context, topic string, key string, data interface{}) error
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

	// Отправляем запрос через Kafka
	request := map[string]interface{}{
		"request_type": "deposit",
		"user_id":      userID,
		"amount":       int64(amount * 100), // конвертируем в копейки
		"operation_id": res.Transaction.ID,
		"timestamp":    time.Now().UTC(),
	}

	err = s.kafkaPublisher.Publish(ctx, "transaction_data", fmt.Sprintf("%d", userID), request)
	if err != nil {
		s.logger.Error().Err(err).Uint64("user_id", userID).Uint64("transaction_id",
			res.Transaction.ID).Msg("failed to send deposit request via kafka")

		// Обновляем статус транзакции на failed
		updateErr := s.repo.UpdateTransactionStatus(ctx, res.Transaction.ID,
			model.TransactionStatusFailed)
		if updateErr != nil {
			s.logger.Error().Err(updateErr).Uint64("transaction_id",
				res.Transaction.ID).Msg("failed to update transaction status to failed")
		}

		return model.TransactionDetails{}, err
	}

	s.logger.Info().Uint64("user_id", userID).Uint64("transaction_id",
		res.Transaction.ID).Msg("deposit request sent via kafka")

	// Транзакция остается в pending-статусе до получения ответа
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

	// Отправляем запрос через Kafka
	request := map[string]interface{}{
		"request_type": "withdraw",
		"user_id":      accountID,
		"amount":       int64(amount * 100), // конвертируем в копейки
		"operation_id": res.Transaction.ID,
		"timestamp":    time.Now().UTC(),
	}

	err = s.kafkaPublisher.Publish(ctx, "transaction_data", fmt.Sprintf("%d", accountID),
		request)
	if err != nil {
		s.logger.Error().Err(err).Uint64("account_id", accountID).Uint64("transaction_id",
			res.Transaction.ID).Msg("failed to send withdraw request via kafka")

		// Обновляем статус транзакции на failed
		updateErr := s.repo.UpdateTransactionStatus(ctx, res.Transaction.ID,
			model.TransactionStatusFailed)

		if updateErr != nil {
			s.logger.Error().Err(updateErr).Uint64("transaction_id",
				res.Transaction.ID).Msg("failed to update transaction status to failed")
		}

		return model.TransactionDetails{}, err
	}

	s.logger.Info().Uint64("account_id", accountID).Uint64("transaction_id",
		res.Transaction.ID).Msg("withdraw request sent via kafka")

	// Транзакция остается в pending-статусе до получения ответа
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

	// Отправляем запрос через Kafka
	request := map[string]interface{}{
		"request_type": "transfer",
		"user_id":      userID,
		"amount":       int64(amount * 100), // конвертируем в копейки
		"operation_id": res.Transaction.ID,
		"recipient_id": recipient,
		"timestamp":    time.Now().UTC(),
	}

	err = s.kafkaPublisher.Publish(ctx, "transaction_data", fmt.Sprintf("%d", userID), request)
	if err != nil {
		s.logger.Error().Err(err).Uint64("user_id", userID).Uint64("recipient",
			recipient).Uint64("transaction_id", res.Transaction.ID).Msg("failed to send transfer request via kafka")

		// Обновляем статус транзакции на failed
		updateErr := s.repo.UpdateTransactionStatus(ctx, res.Transaction.ID,
			model.TransactionStatusFailed)

		if updateErr != nil {
			s.logger.Error().Err(updateErr).Uint64("transaction_id",
				res.Transaction.ID).Msg("failed to update transaction status to failed")
		}

		return model.TransactionDetails{}, err
	}

	s.logger.Info().Uint64("user_id", userID).Uint64("recipient",
		recipient).Uint64("transaction_id", res.Transaction.ID).Msg("transfer request sent via kafka")

	// Транзакция остается в pending-статусе до получения ответа
	return res, nil
}

// HandleAccountResponse обрабатывает ответы от account service через Kafka
func (s *TransactionService) HandleAccountResponse(ctx context.Context, topic string, key string,
	data []byte) error {
	var response AccountResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("failed to unmarshal account response: %w", err)
	}

	if response.RequestType == "" || response.UserID == 0 || response.OperationID == 0 {
		return fmt.Errorf("missing or invalid")
	}

	s.logger.Info().
		Str("request_type", response.RequestType).
		Uint64("user_id", response.UserID).
		Uint64("operation_id", response.OperationID).
		Interface("result", response.Result).
		Msg("received account response via kafka")

	// Обновляем статус транзакции
	var status model.TransactionStatus
	var logMsg string
	if response.Result {
		status = model.TransactionStatusCompleted
		logMsg = "transaction completed successfully"
	} else {
		status = model.TransactionStatusFailed
		logMsg = "transaction failed"
	}

	err := s.repo.UpdateTransactionStatus(ctx, response.OperationID, status)
	if err != nil {
		s.logger.Error().Err(err).Uint64("operation_id", response.OperationID).Msgf("failed to update transaction status to %s", status)
		return fmt.Errorf("failed to update transaction status to %s: %w", status, err)
	}

	s.logger.Info().Uint64("operation_id", response.OperationID).Msg(logMsg)
	return nil
}
