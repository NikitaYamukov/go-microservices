package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/NikitaYamukov/go-microservices/transaction/internal/model"
	"github.com/NikitaYamukov/go-microservices/transaction/internal/repository/mapper"
	repomodel "github.com/NikitaYamukov/go-microservices/transaction/internal/repository/model"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db     *gorm.DB
	logger *zerolog.Logger
}

func NewRepository(db *gorm.DB, logger *zerolog.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) GetTransactions(ctx context.Context, params model.GetTransactionsParams) ([]model.Transaction, error) {
	query := r.db.WithContext(ctx).Model(&repomodel.Transaction{})

	// Применяем фильтры
	if params.UserID != nil {
		query = query.Where("user_id = ?", *params.UserID)
	}
	if params.Type != nil {
		query = query.Where("type = ?", *params.Type)
	}
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	if params.DateFrom != nil {
		query = query.Where("created_at >= ?", *params.DateFrom)
	}
	if params.DateTo != nil {
		query = query.Where("created_at <= ?", *params.DateTo)
	}

	// Применяем пагинацию и сортировку
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}

	var transactions []repomodel.Transaction
	res := query.Order("created_at DESC").Find(&transactions)
	if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to get transactions")
		return nil, fmt.Errorf("failed to get transactions: %w", res.Error)
	}

	return mapper.RepoTransactionsToTransactions(transactions), nil
}

func (r *Repository) GetTransactionDetails(ctx context.Context, transactionID uint64) (model.TransactionDetails, error) {
	// Получаем транзакцию
	var transaction repomodel.Transaction
	res := r.db.WithContext(ctx).
		Model(&repomodel.Transaction{}).
		Where("id = ?", transactionID).
		First(&transaction)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return model.TransactionDetails{}, fmt.Errorf("transaction not found")
	} else if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to get transaction")
		return model.TransactionDetails{}, fmt.Errorf("failed to get transaction: %w", res.Error)
	}

	// Получаем записи транзакции
	var entries []repomodel.TransactionEntry
	res = r.db.WithContext(ctx).
		Model(&repomodel.TransactionEntry{}).
		Where("transaction_id = ?", transactionID).
		Order("created_at ASC").
		Find(&entries)
	if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to get transaction entries")
		return model.TransactionDetails{}, fmt.Errorf("failed to get transaction entries: %w", res.Error)
	}

	return model.TransactionDetails{
		Transaction: mapper.RepoTransactionToTransaction(transaction),
		Entries:     mapper.RepoTransactionEntriesToTransactionEntries(entries),
	}, nil
}

// Deposit выполняет операцию депозита в рамках одной транзакции БД
func (r *Repository) Deposit(ctx context.Context, params model.DepositParams) (model.TransactionDetails, error) {
	var result model.TransactionDetails

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Создаем транзакцию
		transaction := model.Transaction{
			UserID: params.UserID,
			Amount: int64(params.Amount * 100),
			Status: model.TransactionStatusPending,
			Type:   model.TransactionTypeDeposit,
		}

		transactionRepo := mapper.TransactionToRepoTransaction(transaction)
		res := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&transactionRepo)
		if res.Error != nil {
			r.logger.Err(res.Error).Msg("failed to create transaction in deposit")
			return fmt.Errorf("failed to create transaction: %w", res.Error)
		}

		// Создаем запись дебета (увеличение баланса)
		entry := model.TransactionEntry{
			TransactionID: transactionRepo.ID,
			AccountID:     params.UserID,
			Direction:     model.TransactionEntryDirectionDebit,
			Amount:        params.Amount,
		}

		entryRepo := mapper.TransactionEntryToRepoTransactionEntry(entry)
		res = tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&entryRepo)
		if res.Error != nil {
			r.logger.Err(res.Error).Msg("failed to create transaction entry in deposit")
			return fmt.Errorf("failed to create transaction entry: %w", res.Error)
		}

		// Получаем детали созданной транзакции (оставляем в статусе pending)
		transactionRepo.Status = string(model.TransactionStatusPending)
		result = model.TransactionDetails{
			Transaction: mapper.RepoTransactionToTransaction(transactionRepo),
			Entries: []model.TransactionEntry{
				mapper.RepoTransactionEntryToTransactionEntry(entryRepo),
			},
		}

		return nil
	})

	if err != nil {
		return model.TransactionDetails{}, err
	}

	return result, nil
}

// Withdraw выполняет операцию снятия средств в рамках одной транзакции БД
func (r *Repository) Withdraw(ctx context.Context, params model.WithdrawParams) (model.TransactionDetails, error) {
	var result model.TransactionDetails

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Создаем транзакцию
		transaction := model.Transaction{
			UserID: params.AccountID,
			Amount: int64(params.Amount * 100),
			Status: model.TransactionStatusPending,
			Type:   model.TransactionTypeWithdraw,
		}

		transactionRepo := mapper.TransactionToRepoTransaction(transaction)
		res := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&transactionRepo)
		if res.Error != nil {
			r.logger.Err(res.Error).Msg("failed to create transaction in withdraw")
			return fmt.Errorf("failed to create transaction: %w", res.Error)
		}

		// Создаем запись кредита (уменьшение баланса)
		entry := model.TransactionEntry{
			TransactionID: transactionRepo.ID,
			AccountID:     params.AccountID,
			Direction:     model.TransactionEntryDirectionCredit,
			Amount:        params.Amount,
		}

		entryRepo := mapper.TransactionEntryToRepoTransactionEntry(entry)
		res = tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&entryRepo)
		if res.Error != nil {
			r.logger.Err(res.Error).Msg("failed to create transaction entry in withdraw")
			return fmt.Errorf("failed to create transaction entry: %w", res.Error)
		}

		// Получаем детали созданной транзакции (оставляем в статусе pending)
		transactionRepo.Status = string(model.TransactionStatusPending)
		result = model.TransactionDetails{
			Transaction: mapper.RepoTransactionToTransaction(transactionRepo),
			Entries: []model.TransactionEntry{
				mapper.RepoTransactionEntryToTransactionEntry(entryRepo),
			},
		}

		return nil
	})

	if err != nil {
		return model.TransactionDetails{}, err
	}

	return result, nil
}

// Transfer выполняет операцию перевода средств в рамках одной транзакции БД
func (r *Repository) Transfer(ctx context.Context, params model.TransferParams) (model.TransactionDetails, error) {
	var result model.TransactionDetails

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Создаем транзакцию
		transaction := model.Transaction{
			UserID: params.UserID,
			Amount: int64(params.Amount * 100),
			Status: model.TransactionStatusPending,
			Type:   model.TransactionTypeTransfer,
		}

		transactionRepo := mapper.TransactionToRepoTransaction(transaction)
		res := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&transactionRepo)
		if res.Error != nil {
			r.logger.Err(res.Error).Msg("failed to create transaction in transfer")
			return fmt.Errorf("failed to create transaction: %w", res.Error)
		}

		// Создаем запись кредита для счета отправителя (уменьшение баланса)
		creditEntry := model.TransactionEntry{
			TransactionID: transactionRepo.ID,
			AccountID:     params.UserID,
			Direction:     model.TransactionEntryDirectionCredit,
			Amount:        params.Amount,
		}

		creditEntryRepo := mapper.TransactionEntryToRepoTransactionEntry(creditEntry)
		res = tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&creditEntryRepo)
		if res.Error != nil {
			r.logger.Err(res.Error).Msg("failed to create credit transaction entry in transfer")
			return fmt.Errorf("failed to create credit transaction entry: %w", res.Error)
		}

		// Создаем запись дебета для счета получателя (увеличение баланса)
		debitEntry := model.TransactionEntry{
			TransactionID: transactionRepo.ID,
			AccountID:     params.Recipient,
			Direction:     model.TransactionEntryDirectionDebit,
			Amount:        params.Amount,
		}

		debitEntryRepo := mapper.TransactionEntryToRepoTransactionEntry(debitEntry)
		res = tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&debitEntryRepo)
		if res.Error != nil {
			r.logger.Err(res.Error).Msg("failed to create debit transaction entry in transfer")
			return fmt.Errorf("failed to create debit transaction entry: %w", res.Error)
		}

		// Получаем детали созданной транзакции (оставляем в статусе pending)
		transactionRepo.Status = string(model.TransactionStatusPending)
		result = model.TransactionDetails{
			Transaction: mapper.RepoTransactionToTransaction(transactionRepo),
			Entries: []model.TransactionEntry{
				mapper.RepoTransactionEntryToTransactionEntry(creditEntryRepo),
				mapper.RepoTransactionEntryToTransactionEntry(debitEntryRepo),
			},
		}

		return nil
	})

	if err != nil {
		return model.TransactionDetails{}, err
	}

	return result, nil
}

// UpdateTransactionStatus обновляет статус транзакции
func (r *Repository) UpdateTransactionStatus(ctx context.Context, transactionID uint64,
	status model.TransactionStatus) error {
	updateData := mapper.UpdateTransactionToRepoTransaction(model.UpdateTransaction{
		Status: status,
	})

	res := r.db.WithContext(ctx).
		Model(&repomodel.Transaction{}).
		Where("id = ?", transactionID).
		Updates(&updateData)

	if res.Error != nil {
		r.logger.Err(res.Error).Uint64("transaction_id", transactionID).Str("status",
			string(status)).Msg("failed to update transaction status")
		return fmt.Errorf("failed to update transaction status: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("transaction with id %d not found", transactionID)
	}

	return nil
}
