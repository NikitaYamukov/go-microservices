package app

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	accountpb "github.com/NikitaYamukov/contracts/account/go"
	transactionpb "github.com/NikitaYamukov/contracts/transaction/go"
	"github.com/NikitaYamukov/go-microservices/transaction/internal/account"
	"github.com/NikitaYamukov/go-microservices/transaction/internal/config"
	_ "github.com/NikitaYamukov/go-microservices/transaction/internal/migrations"
	"github.com/NikitaYamukov/go-microservices/transaction/internal/repository"
	"github.com/NikitaYamukov/go-microservices/transaction/internal/server"
	"github.com/NikitaYamukov/go-microservices/transaction/internal/service"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/lib/pq"
)

type App struct {
	cfg    *config.Config
	logger zerolog.Logger

	transactionRepository *repository.Repository
	transactionService    *service.TransactionService
	transactionServer     *server.Server
	grpcServer            *grpc.Server

	accountService *account.Service
}

func New(logger zerolog.Logger, cfg *config.Config) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) Run(ctx context.Context) error {
	// Инициализируем gRPC-сервер
	transactionServer, err := a.getTransactionServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to get transaction server: %w", err)
	}

	// Создаем gRPC-сервер
	a.grpcServer = getGRPCServer(transactionServer)

	listenAddr := fmt.Sprintf("%s:%d", a.cfg.Host, a.cfg.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		a.logger.Fatal().Err(err).Msg("failed to listen")
		return err
	}

	a.logger.Info().Msg("gRPC server listening")

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- a.grpcServer.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		a.grpcServer.GracefulStop()
		return ctx.Err()
	case err := <-serverErrCh:
		if err != nil {
			a.logger.Error().Err(err).Msg("failed to serve")
		}
		return err
	}
}

func (a *App) GetRepository(ctx context.Context) (*repository.Repository, error) {
	return a.getRepository(ctx)
}

func (a *App) getRepository(ctx context.Context) (*repository.Repository, error) {
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

func (a *App) getTransactionService(ctx context.Context) (*service.TransactionService, error) {
	if a.transactionService == nil {
		repo, err := a.getRepository(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get repository: %w", err)
		}

		accountSvc, err := a.getAccountService(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get account service: %w", err)
		}

		a.transactionService = service.New(repo, accountSvc, &a.logger)
	}

	return a.transactionService, nil
}

func (a *App) getTransactionServer(ctx context.Context) (*server.Server, error) {
	if a.transactionServer == nil {
		service, err := a.getTransactionService(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get service: %w", err)
		}

		a.transactionServer = server.New(service, &a.logger)
	}

	return a.transactionServer, nil
}

func (a *App) getAccountService(ctx context.Context) (*account.Service, error) {
	if a.accountService == nil {
		// Создаем gRPC-соединение с account service
		conn, err := grpc.Dial(a.cfg.AccountGrpcHost, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("failed to connect to account service: %w", err)
		}

		client := accountpb.NewAccountClient(conn)
		a.accountService = account.New(client)
	}

	return a.accountService, nil
}

func getGRPCServer(srv *server.Server) *grpc.Server {
	grpcSrv := grpc.NewServer()
	transactionpb.RegisterTransactionServiceServer(grpcSrv, srv)
	return grpcSrv
}

// Close корректно завершает работу приложения
func (a *App) Close() error {
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}

	return nil
}
