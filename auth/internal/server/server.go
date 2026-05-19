package server

import (
	"context"

	authpb "github.com/NikitaYamukov/contracts/auth/go"
	"github.com/NikitaYamukov/go-microservices/auth/internal/model"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	authpb.UnimplementedAuthServer

	authService AuthService
	logger      *zerolog.Logger
}

func New(authService AuthService, logger *zerolog.Logger) *Server {
	return &Server{authService: authService, logger: logger}
}

type AuthService interface {
	Register(ctx context.Context, req model.Register) error
	Login(ctx context.Context, req model.Login) (model.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (model.TokenPair, error)
	Validate(ctx context.Context, accessToken string) (uint64, error)
	Logout(ctx context.Context, refreshToken string) error
}

func (s *Server) Register(ctx context.Context, req *authpb.RegisterRequest) (*emptypb.Empty, error) {
	err := s.authService.Register(ctx, model.Register{
		Login:    req.GetLogin(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.TokenPair, error) {
	tokens, err := s.authService.Login(ctx, model.Login{
		LoginOrEmail: req.GetLoginOrEmail(),
		Password:     req.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	return tokenPairToPb(tokens), nil
}

func (s *Server) Refresh(ctx context.Context, req *authpb.RefreshRequest) (*authpb.TokenPair, error) {
	tokens, err := s.authService.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return tokenPairToPb(tokens), nil
}

func (s *Server) Validate(ctx context.Context, req *authpb.ValidateRequest) (*authpb.ValidateResponse, error) {
	userID, err := s.authService.Validate(ctx, req.GetAccessToken())
	if err != nil {
		return nil, err
	}

	return &authpb.ValidateResponse{UserId: userID}, nil
}

func (s *Server) Logout(ctx context.Context, req *authpb.RefreshRequest) (*emptypb.Empty, error) {
	if err := s.authService.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func tokenPairToPb(tokens model.TokenPair) *authpb.TokenPair {
	return &authpb.TokenPair{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}
}
