package mapper

import (
	"github.com/NikitaYamukov/go-microservices/auth/internal/model"
	repomodel "github.com/NikitaYamukov/go-microservices/auth/internal/repository/model"
)

func UserToRepoUser(user model.User) repomodel.User {
	return repomodel.User{
		ID:           user.ID,
		Login:        user.Login,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func RepoUserToUser(user repomodel.User) model.User {
	return model.User{
		ID:           user.ID,
		Login:        user.Login,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func RefreshTokenToRepoRefresh(token model.RefreshToken) repomodel.RefreshToken {
	return repomodel.RefreshToken{
		ID:        token.ID,
		UserID:    token.UserID,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
		RevokedAt: token.RevokedAt,
		CreatedAt: token.CreatedAt,
	}
}

func RepoRefreshTokenToRefresh(token repomodel.RefreshToken) model.RefreshToken {
	return model.RefreshToken{
		ID:        token.ID,
		UserID:    token.UserID,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
		RevokedAt: token.RevokedAt,
		CreatedAt: token.CreatedAt,
	}
}
