package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/NikitaYamukov/go-microservices/transaction/internal/config"
	_ "github.com/NikitaYamukov/go-microservices/transaction/internal/migrations"
	"github.com/NikitaYamukov/go-microservices/transaction/internal/repository"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/lib/pq"
)

type App struct {
	cfg    *config.Config
	logger zerolog.Logger

	transactionRepository *repository.Repository
}

func New(logger zerolog.Logger, cfg *config.Config) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) GetRepository(ctx context.Context) (*repository.Repository, error) {
	if a.transactionRepository == nil {
		if err := a.runMigrations(ctx); err != nil {
			return nil, fmt.Errorf("failed to get repository: %w", err)
		}

		db, err := gorm.Open(postgres.Open(a.cfg.DbDsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("gorm init failed: %w", err)
		}

		a.transactionRepository = repository.NewRepository(db, &a.logger)
	}

	return a.transactionRepository, nil
}

func (a *App) runMigrations(ctx context.Context) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to select migrations dialect: %w", err)
	}

	dbGoose, err := sql.Open("postgres", a.cfg.DbDsn)
	if err != nil {
		return fmt.Errorf("failed to create sql connection: %w", err)
	}
	defer dbGoose.Close()

	if err := goose.UpContext(ctx, dbGoose, "internal/migrations"); err != nil {
		return fmt.Errorf("failed to run up migrations: %w", err)
	}

	return nil
}
