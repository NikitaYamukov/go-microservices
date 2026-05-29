package mapper

import (
	"time"

	accountpb "github.com/NikitaYamukov/contracts/account/go"
	authpb "github.com/NikitaYamukov/contracts/auth/go"
	"github.com/NikitaYamukov/go-microservices/gateway/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

func UserToPb(user model.User) *accountpb.User {
	return &accountpb.User{
		Id:         user.ID,
		Login:      user.Login,
		Email:      user.Email,
		Phone:      user.Phone,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Age:        user.Age,
		CreatedAt:  timestamppb.New(user.CreatedAt),
		UpdatedAt:  timestamppb.New(user.UpdatedAt),
	}
}

func UsersToPbs(users []model.User) []*accountpb.User {
	pbs := make([]*accountpb.User, len(users))
	for i, user := range users {
		pbs[i] = UserToPb(user)
	}
	return pbs
}

func PbsToUsers(pbs []*accountpb.User) []model.User {
	users := make([]model.User, len(pbs))
	for i, pb := range pbs {
		users[i] = PbToUser(pb)
	}
	return users
}

func PbToUserCreate(accountpbUser *accountpb.CreateUser) model.CreateUser {
	return model.CreateUser{
		Login:      accountpbUser.Login,
		Email:      accountpbUser.Email,
		Phone:      accountpbUser.Phone,
		FirstName:  accountpbUser.FirstName,
		LastName:   accountpbUser.LastName,
		MiddleName: accountpbUser.MiddleName,
		Age:        accountpbUser.Age,
	}
}

func PbToUserUpdate(userpb *accountpb.User) model.UpdateUser {
	return model.UpdateUser{
		Email:      userpb.Email,
		Phone:      userpb.Phone,
		FirstName:  userpb.FirstName,
		LastName:   userpb.LastName,
		MiddleName: userpb.MiddleName,
		Age:        userpb.Age,
	}
}

func TokenPairToPb(tokenPair model.TokenPair) *authpb.TokenPair {
	return &authpb.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}
}

func PbToTokenPair(pb *authpb.TokenPair) model.TokenPair {
	return model.TokenPair{
		AccessToken:  pb.AccessToken,
		RefreshToken: pb.RefreshToken,
	}
}

func CreateUserToUser(user model.CreateUser) model.User {
	return model.User{
		Login:      user.Login,
		Email:      user.Email,
		Phone:      user.Phone,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Age:        user.Age,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}
