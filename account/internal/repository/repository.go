package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/NikitaYamukov/go-microservices/account/internal/model"
	"github.com/NikitaYamukov/go-microservices/account/internal/repository/mapper"
	repomodel "github.com/NikitaYamukov/go-microservices/account/internal/repository/model"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db     *gorm.DB
	logger zerolog.Logger
}

func NewRepository(db *gorm.DB, logger zerolog.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	userRepo := mapper.UserToRepoUser(user)
	fmt.Println(userRepo)
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		//Clauses(clause.Returning{}).
		Create(&userRepo)
	if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to save user")
		return model.User{}, fmt.Errorf("failed to save user: %w", res.Error)
	}

	fmt.Println(userRepo)
	return mapper.RepoUserToUser(userRepo), nil
}

func (r *Repository) GetUser(ctx context.Context, userID uint64) (model.User, error) {
	var user repomodel.User
	res := r.db.WithContext(ctx).
		Model(&repomodel.User{}).
		Where("id = ? AND is_deleted = ?", userID, false).
		First(&user)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return model.User{}, fmt.Errorf("user not found")
	} else if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to get user")
		return model.User{}, fmt.Errorf("failed to get user: %w", res.Error)
	}

	return mapper.RepoUserToUser(user), nil
}

func (r *Repository) GetUsers(ctx context.Context, limit int, offset int) ([]model.User, error) {
	var users []repomodel.User
	res := r.db.WithContext(ctx).
		Model(&repomodel.User{}).
		Where("is_deleted = ?", false).
		Offset(offset).
		Limit(limit).
		Find(&users)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("users not found")
	} else if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to get users")
		return nil, fmt.Errorf("failed to get users: %w", res.Error)
	}

	return mapper.RepoUsersToUsers(users), nil
}

func (r *Repository) DeleteUser(ctx context.Context, userID uint64) error {
	res := r.db.WithContext(ctx).
		Model(&repomodel.User{}).
		Where("id = ?", userID).
		Update("is_deleted", true)
	if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to delete user")
		return fmt.Errorf("failed to delete user: %w", res.Error)
	}

	return nil
}

// GetBalance получает текущий баланс пользователя
func (r *Repository) GetBalance(ctx context.Context, userID uint64) (float32, error) {
	var user repomodel.User
	res := r.db.WithContext(ctx).
		Model(&repomodel.User{}).
		Select("balance").
		Where("id = ? AND is_deleted = ?", userID, false).
		First(&user)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("user not found")
	} else if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to get user balance")
		return 0, fmt.Errorf("failed to get user balance: %w", res.Error)
	}

	return user.Balance, nil
}

// UpdateBalance обновляет баланс пользователя
func (r *Repository) UpdateBalance(ctx context.Context, userID uint64, amount float32,
	operationType model.BalanceOperationType) (float32, float32, error) {
	var user repomodel.User

	// Сначала получаем текущий баланс
	res := r.db.WithContext(ctx).
		Model(&repomodel.User{}).
		Where("id = ? AND is_deleted = ?", userID, false).
		First(&user)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return 0, 0, fmt.Errorf("user not found")
	} else if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to get user for balance update")
		return 0, 0, fmt.Errorf("failed to get user for balance update: %w", res.Error)
	}

	oldBalance := user.Balance
	var newBalance float32

	// Вычисляем новый баланс в зависимости от типа операции
	switch operationType {
	case model.BalanceOperationDeposit:
		newBalance = oldBalance + amount
	case model.BalanceOperationCredit:
		newBalance = oldBalance - amount
		if newBalance < 0 {
			return 0, 0, fmt.Errorf("insufficient balance: current balance %.2f, requested amount %.2f",
				oldBalance, amount)
		}
	default:
		return 0, 0, fmt.Errorf("invalid balance operation type: %s", operationType)
	}

	// Обновляем баланс в базе данных
	res = r.db.WithContext(ctx).
		Model(&repomodel.User{}).
		Where("id = ? AND is_deleted = ?", userID, false).
		Update("balance", newBalance)

	if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to update user balance")
		return 0, 0, fmt.Errorf("failed to update user balance: %w", res.Error)
	}

	return oldBalance, newBalance, nil
}

// TransferBalance переводит средства между пользователями
func (r *Repository) TransferBalance(ctx context.Context, fromUserID, toUserID uint64, amount float32) error {
	if amount <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	if fromUserID == toUserID {
		return fmt.Errorf("cannot transfer to the same user")
	}

	// Начинаем транзакцию
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Проверяем существование получателя
		var toUser repomodel.User
		res := tx.Model(&repomodel.User{}).
			Where("id = ? AND is_deleted = ?", toUserID, false).
			First(&toUser)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("recipient user not found")
		} else if res.Error != nil {
			return fmt.Errorf("failed to get recipient user: %w", res.Error)
		}

		// Получаем текущий баланс отправителя
		var fromUser repomodel.User
		res = tx.Model(&repomodel.User{}).
			Where("id = ? AND is_deleted = ?", fromUserID, false).
			First(&fromUser)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("sender user not found")
		} else if res.Error != nil {
			return fmt.Errorf("failed to get sender user: %w", res.Error)
		}

		// Проверяем достаточность средств
		if fromUser.Balance < amount {
			return fmt.Errorf("insufficient balance: current balance %.2f, transfer amount %.2f",
				fromUser.Balance, amount)
		}

		// Списываем средства с отправителя
		res = tx.Model(&repomodel.User{}).
			Where("id = ? AND is_deleted = ?", fromUserID, false).
			Update("balance", fromUser.Balance-amount)
		if res.Error != nil {
			return fmt.Errorf("failed to deduct from sender balance: %w", res.Error)
		}

		// Зачисляем средства получателю
		res = tx.Model(&repomodel.User{}).
			Where("id = ? AND is_deleted = ?", toUserID, false).
			Update("balance", toUser.Balance+amount)
		if res.Error != nil {
			return fmt.Errorf("failed to add to recipient balance: %w", res.Error)
		}

		return nil
	})
}

func (r *Repository) UpdateUser(ctx context.Context, userID uint64, user model.UpdateUser) error {
	userRepo := mapper.UpdateUserToRepoUser(user)
	res := r.db.WithContext(ctx).
		Model(&repomodel.User{}).
		Where("id = ? AND is_deleted = ?", userID, false).
		Updates(userRepo)
	if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to update user")
		return fmt.Errorf("failed to update user: %w", res.Error)
	}

	return nil
}
