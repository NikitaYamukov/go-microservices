package server

import (
	"context"

	accountpb "github.com/NikitaYamukov/contracts/account/go"
	"github.com/NikitaYamukov/go-microservices/account/internal/mapper"
	"github.com/NikitaYamukov/go-microservices/account/internal/model"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	accountpb.UnimplementedAccountServer

	accountService AccountService
	logger         zerolog.Logger
}

func New(accountService AccountService, logger zerolog.Logger) *Server {
	return &Server{accountService: accountService, logger: logger}
}

type AccountService interface {
	CreateUser(ctx context.Context, user model.CreateUser) (model.User, error)
	GetUser(ctx context.Context, userID uint64) (model.User, error)
	GetUsers(ctx context.Context, limit int, offset int) ([]model.User, error)
	DeleteUser(ctx context.Context, userID uint64) error
	UpdateUser(ctx context.Context, userID uint64, user model.UpdateUser) error
	GetBalance(ctx context.Context, userID uint64) (model.GetBalanceResponse, error)
	Deposit(ctx context.Context, userID uint64, amount int64) (model.UpdateBalanceResponse, error)
	Withdraw(ctx context.Context, userID uint64, amount int64) (model.UpdateBalanceResponse, error)
	Transfer(ctx context.Context, fromUserID, toUserID uint64, amount int64) (model.GetBalanceResponse,
		model.GetBalanceResponse, error)
}

func (s *Server) CreateUser(ctx context.Context, req *accountpb.CreateUserRequest) (*accountpb.CreateUserResponse, error) {
	user := mapper.PbToUserCreate(req.GetUser())
	res, err := s.accountService.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &accountpb.CreateUserResponse{User: mapper.UserToPb(res)}, nil
}

func (s *Server) GetUser(ctx context.Context, req *accountpb.GetUserRequest) (*accountpb.GetUserResponse, error) {
	res, err := s.accountService.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &accountpb.GetUserResponse{
		User: mapper.UserToPb(res),
	}, nil
}

func (s *Server) GetUsers(ctx context.Context, req *accountpb.GetUsersRequest) (*accountpb.GetUsersResponse, error) {
	res, err := s.accountService.GetUsers(ctx, int(req.GetPagination().GetLimit()), int(req.GetPagination().GetOffset()))
	if err != nil {
		return nil, err
	}

	return &accountpb.GetUsersResponse{
		Users: mapper.UsersToPbs(res),
	}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *accountpb.UpdateUserRequest) (*emptypb.Empty, error) {
	user := mapper.PbToUserUpdate(req.GetUser())
	if err := s.accountService.UpdateUser(ctx, req.GetUserId(), user); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *accountpb.DeleteUserRequest) (*emptypb.Empty, error) {
	if err := s.accountService.DeleteUser(ctx, req.GetUserId()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Deposit пополняет баланс пользователя
func (s *Server) Deposit(ctx context.Context, req *accountpb.DepositRequest) (*accountpb.DepositResponse, error) {
	response, err := s.accountService.Deposit(ctx, req.GetUserId(), req.GetAmount())
	if err != nil {
		return &accountpb.DepositResponse{
			Status:  "failed",
			Balance: 0,
		}, err
	}

	return &accountpb.DepositResponse{
		Status:  "completed",
		Balance: int64(response.NewBalance),
	}, nil
}

// Withdraw списывает средства с баланса пользователя
func (s *Server) Withdraw(ctx context.Context, req *accountpb.WithdrawRequest) (*accountpb.WithdrawResponse, error) {
	response, err := s.accountService.Withdraw(ctx, req.GetUserId(), req.GetAmount())
	if err != nil {
		return &accountpb.WithdrawResponse{
			Status:  "failed",
			Balance: 0,
		}, err
	}

	return &accountpb.WithdrawResponse{
		Status:  "completed",
		Balance: int64(response.NewBalance),
	}, nil
}

// Transfer переводит средства между пользователями
func (s *Server) Transfer(ctx context.Context, req *accountpb.TransferRequest) (*accountpb.TransferResponse, error) {
	fromUserBalance, toUserBalance, err := s.accountService.Transfer(ctx, req.GetUserId(),
		req.GetRecipientId(), req.GetAmount())
	if err != nil {
		return &accountpb.TransferResponse{
			Status:           "failed",
			UserBalance:      0,
			RecipientBalance: 0,
		}, err
	}

	return &accountpb.TransferResponse{
		Status:           "completed",
		UserBalance:      int64(fromUserBalance.Balance),
		RecipientBalance: int64(toUserBalance.Balance),
	}, nil
}

// GetBalance получает баланс пользователя
func (s *Server) GetBalance(ctx context.Context, req *accountpb.GetBalanceRequest) (*accountpb.GetBalanceResponse, error) {
	response, err := s.accountService.GetBalance(ctx, req.GetUserId())
	if err != nil {
		return &accountpb.GetBalanceResponse{
			Balance: 0,
		}, err
	}

	return &accountpb.GetBalanceResponse{
		Balance: int64(response.Balance),
	}, nil
}
