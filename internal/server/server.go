package server

import (
	"context"

	accountpb "github.com/NikitaYamukov/contracts/account/go"
	"github.com/NikitaYamukov/go-microservices/internal/mapper"
	"github.com/NikitaYamukov/go-microservices/internal/model"
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
	CreateUser(ctx context.Context, user model.CreateUser) error
	GetUser(ctx context.Context, userID uint64) (model.User, error)
	GetUsers(ctx context.Context, limit int, offset int) ([]model.User, error)
	DeleteUser(ctx context.Context, userID uint64) error
	UpdateUser(ctx context.Context, userID uint64, user model.UpdateUser) error
}

func (s *Server) CreateUser(ctx context.Context, req *accountpb.CreateUserRequest) (*emptypb.Empty, error) {
	user := mapper.PbToUserCreate(req.GetUser())
	if err := s.accountService.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
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
