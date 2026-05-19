package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/NikitaYamukov/go-microservices/auth/internal/model"
	"github.com/NikitaYamukov/go-microservices/auth/internal/repository/mapper"
	repomodel "github.com/NikitaYamukov/go-microservices/auth/internal/repository/model"
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

func (r *Repository) CreateUser(ctx context.Context, user model.User) error {
	userRepo := mapper.UserToRepoUser(user)
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&userRepo)

	if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to save user")
		return fmt.Errorf("failed to save user: %w", res.Error)
	}

	return nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID uint64) (model.User, error) {
	var user repomodel.User
	res := r.db.WithContext(ctx).
		Model(&repomodel.User{}).
		Where("id = ?", userID).
		First(&user)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return model.User{}, fmt.Errorf("user not found")
	} else if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to get user id")
		return model.User{}, res.Error
	}

	return mapper.RepoUserToUser(user), nil
}

func (r *Repository) GetUserByLoginOrEmail(ctx context.Context, loginOrEmail string) (model.User, error) {
	var user repomodel.User
	res := r.db.WithContext(ctx).
		Model(&repomodel.User{}).
		Where("login = ? OR email = ?", loginOrEmail, loginOrEmail).
		First(&user)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return model.User{}, fmt.Errorf("user not found")
	} else if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to get user by login/email")
		return model.User{}, res.Error
	}

	return mapper.RepoUserToUser(user), nil
}

func (r *Repository) SaveRefreshToken(ctx context.Context, token model.RefreshToken) error {
	m := mapper.RefreshTokenToRepoRefresh(token)
	res := r.db.WithContext(ctx).Create(&m)

	if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to save refresh token")
		return fmt.Errorf("failed to save refresh token: %w", res.Error)
	}

	return nil
}

func (r *Repository) GetRefreshToken(ctx context.Context, token string) (model.RefreshToken, error) {
	var refreshToken repomodel.RefreshToken
	res := r.db.WithContext(ctx).
		Model(&repomodel.RefreshToken{}).
		Where("token = ?", token).
		First(&refreshToken)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return model.RefreshToken{}, fmt.Errorf("refresh token not found")
	} else if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to get refresh token")
		return model.RefreshToken{}, res.Error
	}

	return mapper.RepoRefreshTokenToRefresh(refreshToken), nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, token string) error {
	res := r.db.WithContext(ctx).
		Model(&repomodel.RefreshToken{}).
		Where("token = ?", token).
		Update("revoked_at", gorm.Expr("NOW()"))

	if res.Error != nil {
		r.logger.Err(res.Error).Msg("failed to revoke refresh token")
		return fmt.Errorf("failed to revoke refresh token")
	}

	return nil
}
