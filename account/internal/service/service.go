package service

import (
	"context"
	"fmt"
	"time"

	"github.com/NikitaYamukov/go-microservices/account/internal/model"
	repomapper "github.com/NikitaYamukov/go-microservices/account/internal/repository/mapper"
	"github.com/rs/zerolog"
)

type AccountService struct {
	repo   Repository
	logger zerolog.Logger
}

func New(repo Repository, logger zerolog.Logger) *AccountService {
	return &AccountService{repo: repo, logger: logger}
}

type Repository interface {
	CreateUser(context.Context, model.User) (model.User, error)
	GetUser(context.Context, uint64) (model.User, error)
	GetUsers(context.Context, int, int) ([]model.User, error)
	DeleteUser(context.Context, uint64) error
	UpdateUser(context.Context, uint64, model.UpdateUser) error
	GetBalance(context.Context, uint64) (float32, error)
	UpdateBalance(context.Context, uint64, float32, model.BalanceOperationType) (float32, float32, error)
	TransferBalance(context.Context, uint64, uint64, float32) error
}

func (s *AccountService) CreateUser(ctx context.Context, newUser model.CreateUser) (model.User, error) {
	repoUser := repomapper.CreateUserToRepoUser(newUser)
	user := repomapper.RepoUserToUser(repoUser)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return s.repo.CreateUser(ctx, user)
}

func (s *AccountService) GetUsers(ctx context.Context, limit int, offset int) ([]model.User, error) {
	return s.repo.GetUsers(ctx, limit, offset)
}

func (s *AccountService) GetUser(ctx context.Context, userID uint64) (model.User, error) {
	return s.repo.GetUser(ctx, userID)
}

func (s *AccountService) DeleteUser(ctx context.Context, userID uint64) error {
	return s.repo.DeleteUser(ctx, userID)
}

func (s *AccountService) UpdateUser(ctx context.Context, userID uint64, user model.UpdateUser) error {
	return s.repo.UpdateUser(ctx, userID, user)
}

// GetBalance получает текущий баланс пользователя
func (s *AccountService) GetBalance(ctx context.Context, userID uint64) (model.GetBalanceResponse, error) {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Err(err).Uint64("userID", userID).Msg("failed to get user balance")
		return model.GetBalanceResponse{}, err
	}

	return model.GetBalanceResponse{
		UserID:  userID,
		Balance: balance,
	}, nil
}

// UpdateBalance обновляет баланс пользователя
func (s *AccountService) UpdateBalance(ctx context.Context, req model.UpdateBalanceRequest) (
	model.UpdateBalanceResponse, error) {
	// Валидация входных данных
	if req.Amount < 0 {
		return model.UpdateBalanceResponse{}, fmt.Errorf("amount cannot be negative")
	}

	if req.Type == model.BalanceOperationCredit && req.Amount == 0 {
		return model.UpdateBalanceResponse{}, fmt.Errorf("amount must be positive for subtract operation")
	}

	oldBalance, newBalance, err := s.repo.UpdateBalance(ctx, req.UserID, req.Amount, req.Type)
	if err != nil {
		s.logger.Err(err).
			Uint64("userID", req.UserID).
			Float32("amount", req.Amount).
			Str("operation", string(req.Type)).
			Msg("failed to update user balance")
		return model.UpdateBalanceResponse{}, err
	}

	s.logger.Info().
		Uint64("userID", req.UserID).
		Float32("oldBalance", oldBalance).
		Float32("newBalance", newBalance).
		Float32("amount", req.Amount).
		Str("operation", string(req.Type)).
		Msg("user balance updated successfully")

	return model.UpdateBalanceResponse{
		UserID:     req.UserID,
		OldBalance: oldBalance,
		NewBalance: newBalance,
		Amount:     req.Amount,
		Operation:  req.Type,
	}, nil
}

// TransferBalance переводит средства между пользователями
func (s *AccountService) TransferBalance(ctx context.Context, fromUserID, toUserID uint64, amount float32) error {
	if amount <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	if fromUserID == toUserID {
		return fmt.Errorf("cannot transfer to the same user")
	}

	err := s.repo.TransferBalance(ctx, fromUserID, toUserID, amount)
	if err != nil {
		s.logger.Err(err).
			Uint64("fromUserID", fromUserID).
			Uint64("toUserID", toUserID).
			Float32("amount", amount).
			Msg("failed to transfer balance")
		return err
	}

	s.logger.Info().
		Uint64("fromUserID", fromUserID).
		Uint64("toUserID", toUserID).
		Float32("amount", amount).
		Msg("balance transferred successfully")

	return nil
}

// Deposit пополняет баланс пользователя
func (s *AccountService) Deposit(ctx context.Context, userID uint64, amount int64) (
	model.UpdateBalanceResponse, error) {
	updateReq := model.UpdateBalanceRequest{
		UserID: userID,
		Amount: float32(amount),
		Type:   model.BalanceOperationDeposit,
	}

	response, err := s.UpdateBalance(ctx, updateReq)
	if err != nil {
		s.logger.Err(err).
			Uint64("userID", userID).
			Int64("amount", amount).
			Msg("failed to deposit")
		return model.UpdateBalanceResponse{}, err
	}

	s.logger.Info().
		Uint64("userID", userID).
		Int64("amount", amount).
		Float32("newBalance", response.NewBalance).
		Msg("deposit successful")

	return response, nil
}

// Withdraw списывает средства с баланса пользователя
func (s *AccountService) Withdraw(ctx context.Context, userID uint64, amount int64) (
	model.UpdateBalanceResponse, error) {
	updateReq := model.UpdateBalanceRequest{
		UserID: userID,
		Amount: float32(amount),
		Type:   model.BalanceOperationCredit,
	}

	response, err := s.UpdateBalance(ctx, updateReq)
	if err != nil {
		s.logger.Err(err).
			Uint64("userID", userID).
			Int64("amount", amount).
			Msg("failed to withdraw")
		return model.UpdateBalanceResponse{}, err
	}

	s.logger.Info().
		Uint64("userID", userID).
		Int64("amount", amount).
		Float32("newBalance", response.NewBalance).
		Msg("withdraw successful")

	return response, nil
}

// Transfer переводит средства между пользователями и возвращает балансы
func (s *AccountService) Transfer(ctx context.Context, fromUserID, toUserID uint64, amount int64) (
	model.GetBalanceResponse, model.GetBalanceResponse, error) {
	err := s.TransferBalance(ctx, fromUserID, toUserID, float32(amount))
	if err != nil {
		s.logger.Err(err).
			Uint64("fromUserID", fromUserID).
			Uint64("toUserID", toUserID).
			Int64("amount", amount).
			Msg("failed to transfer")
		return model.GetBalanceResponse{}, model.GetBalanceResponse{}, err
	}

	// Получаем балансы обоих пользователей для ответа
	fromUserBalance, err := s.GetBalance(ctx, fromUserID)
	if err != nil {
		s.logger.Err(err).Uint64("userID", fromUserID).Msg("failed to get sender balance")
		return model.GetBalanceResponse{}, model.GetBalanceResponse{}, err
	}

	toUserBalance, err := s.GetBalance(ctx, toUserID)
	if err != nil {
		s.logger.Err(err).Uint64("userID", toUserID).Msg("failed to get recipient balance")
		return model.GetBalanceResponse{}, model.GetBalanceResponse{}, err
	}

	s.logger.Info().
		Uint64("fromUserID", fromUserID).
		Uint64("toUserID", toUserID).
		Int64("amount", amount).
		Msg("transfer successful")

	return fromUserBalance, toUserBalance, nil
}
