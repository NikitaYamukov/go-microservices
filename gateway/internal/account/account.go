package account

import (
	"context"
	"fmt"

	accountpb "github.com/NikitaYamukov/contracts/account/go"
	pagination "github.com/NikitaYamukov/contracts/pagination/go"
	"github.com/NikitaYamukov/go-microservices/gateway/internal/model"
)

type Service struct {
	client accountpb.AccountClient
}

func New(client accountpb.AccountClient) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) GetUser(ctx context.Context, userID uint64) (model.User, error) {
	user, err := s.client.GetUser(ctx, &accountpb.GetUserRequest{UserId: userID})
	if err != nil {
		return model.User{}, fmt.Errorf("failed get user: %w", err)
	}

	return PbToUser(user.User), nil
}

func (s *Service) GetUsers(ctx context.Context, limit uint32, offset uint32) ([]model.User, error) {
	users, err := s.client.GetUsers(ctx, &accountpb.GetUsersRequest{Pagination: &pagination.Pagination{
		Limit:  limit,
		Offset: offset,
	}})
	if err != nil {
		return nil, fmt.Errorf("failed get users: %w", err)
	}

	return PbsToUsers(users.Users), nil
}

func (s *Service) DeleteUser(ctx context.Context, userID uint64) error {
	_, err := s.client.DeleteUser(ctx, &accountpb.DeleteUserRequest{UserId: userID})
	if err != nil {
		return fmt.Errorf("failed delete user: %w", err)
	}

	return nil
}

func (s *Service) CreateUser(ctx context.Context, user model.User) error {
	_, err := s.client.CreateUser(ctx, &accountpb.CreateUserRequest{
		User: UserCreateToPb(user),
	})
	if err != nil {
		return fmt.Errorf("failed create user: %w", err)
	}

	return nil
}

func (s *Service) UpdateUser(ctx context.Context, userID uint64, user model.UpdateUser) error {
	_, err := s.client.UpdateUser(ctx, &accountpb.UpdateUserRequest{
		UserId: userID,
		User:   UserUpdateToPb(user),
	})
	if err != nil {
		return fmt.Errorf("failed update user: %w", err)
	}

	return nil
}

func PbsToUsers(pbs []*accountpb.User) []model.User {
	users := make([]model.User, len(pbs))
	for i, pb := range pbs {
		users[i] = PbToUser(pb)
	}

	return users
}

func PbToUser(userpb *accountpb.User) model.User {
	return model.User{
		ID:         userpb.Id,
		Login:      userpb.Login,
		Email:      userpb.Email,
		Phone:      userpb.Phone,
		FirstName:  userpb.FirstName,
		LastName:   userpb.LastName,
		MiddleName: userpb.MiddleName,
		Age:        userpb.Age,
		CreatedAt:  userpb.CreatedAt.AsTime(),
		UpdatedAt:  userpb.UpdatedAt.AsTime(),
	}
}

func UserCreateToPb(user model.User) *accountpb.CreateUser {
	return &accountpb.CreateUser{
		Login:      user.Login,
		Email:      user.Email,
		Phone:      user.Phone,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Age:        user.Age,
	}
}

func UserUpdateToPb(user model.UpdateUser) *accountpb.User {
	return &accountpb.User{
		Email:      user.Email,
		Phone:      user.Phone,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Age:        user.Age,
	}
}
