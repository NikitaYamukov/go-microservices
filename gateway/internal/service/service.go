package service

import (
	"context"
	"fmt"

	"github.com/NikitaYamukov/go-microservices/gateway/internal/mapper"
	"github.com/NikitaYamukov/go-microservices/gateway/internal/model"
	"github.com/rs/zerolog"
)

type GatewayService struct {
	logger *zerolog.Logger

	accountService AccountService
	authService    AuthService
}

func New(accountService AccountService, authService AuthService, logger *zerolog.Logger) *GatewayService {
	return &GatewayService{
		accountService: accountService,
		authService:    authService,
		logger:         logger,
	}
}

type AccountService interface {
	CreateUser(context.Context, model.User) error
	GetUser(context.Context, uint64) (model.User, error)
	GetUsers(context.Context, uint32, uint32) ([]model.User, error)
	DeleteUser(context.Context, uint64) error
	UpdateUser(context.Context, uint64, model.UpdateUser) error
}

type AuthService interface {
	Login(context.Context, string, string) (model.TokenPair, error)
	Logout(context.Context, string) error
	Register(context.Context, string, string, string) error
	Verify(context.Context, string) (uint64, error)
	Refresh(context.Context, string) (model.TokenPair, error)
}

// Методы для аутентификации.
func (s *GatewayService) Register(ctx context.Context, newUser model.User, password string) (model.User, model.TokenPair, error) {
	if err := s.accountService.CreateUser(ctx, newUser); err != nil {
		return model.User{}, model.TokenPair{}, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.authService.Register(ctx, newUser.Login, newUser.Email, password); err != nil {
		return model.User{}, model.TokenPair{}, fmt.Errorf("failed to register user: %w", err)
	}

	token, err := s.authService.Login(ctx, newUser.Login, password)
	if err != nil {
		return model.User{}, model.TokenPair{}, fmt.Errorf("failed to login: %w", err)
	}

	return newUser, token, nil
}

func (s *GatewayService) Login(ctx context.Context, loginOrEmail, password string) (model.User, model.TokenPair, error) {
	tokens, err := s.authService.Login(ctx, loginOrEmail, password)
	if err != nil {
		return model.User{}, model.TokenPair{}, fmt.Errorf("user not found: %w", err)
	}
	user := model.User{}

	return user, tokens, nil
}

func (s *GatewayService) Logout(ctx context.Context, token string) error {
	if err := s.authService.Logout(ctx, token); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}

func (s *GatewayService) Refresh(ctx context.Context, refreshToken string) (model.TokenPair, error) {
	return s.authService.Refresh(ctx, refreshToken)
}

func (s *GatewayService) ValidateToken(ctx context.Context, accessToken string) (uint64, bool, error) {
	userID, err := s.authService.Verify(ctx, accessToken)
	if err != nil {
		return 0, false, err
	}

	return userID, true, nil
}

// Методы для управления пользователями.
func (s *GatewayService) CreateUser(ctx context.Context, newUser model.CreateUser) error {
	user := mapper.CreateUserToUser(newUser)

	if err := s.accountService.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("failed create user: %w", err)
	}

	return nil
}

func (s *GatewayService) GetUsers(ctx context.Context, limit int, offset int) ([]model.User, error) {
	return s.accountService.GetUsers(ctx, uint32(limit), uint32(offset))
}

func (s *GatewayService) GetUser(ctx context.Context, userID uint64) (model.User, error) {
	return s.accountService.GetUser(ctx, userID)
}

func (s *GatewayService) DeleteUser(ctx context.Context, userID uint64) error {
	if err := s.accountService.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *GatewayService) UpdateUser(ctx context.Context, userID uint64, user model.UpdateUser) error {
	return s.accountService.UpdateUser(ctx, userID, user)
}
