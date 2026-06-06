package app

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	accountpb "github.com/NikitaYamukov/contracts/account/go"
	transactionpb "github.com/NikitaYamukov/contracts/transaction/go"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"transaction/internal/account"
	"transaction/internal/config"
	"transaction/internal/kafka"
	_ "transaction/internal/migrations"
	"transaction/internal/repository"
	"transaction/internal/server"
	"transaction/internal/service"

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
	kafkaClient    *kafka.Kafka
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

		// Получаем gRPC-сервис
		grpcSvc, err := a.getAccountService(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get account service: %w", err)
		}

		// Теперь получаем Kafka-клиент и подписываемся
		kafkaClient, err := a.getKafkaClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get kafka client: %w", err)
		}

		// Обновляем сервис с Kafka-клиентом
		a.transactionService = service.New(repo, grpcSvc, kafkaClient, &a.logger)

		a.kafkaClient.Subscribe(ctx, "transaction_response",
			a.transactionService.HandleAccountResponse)
	}

	return a.transactionService, nil
}

func (a *App) getKafkaClient(ctx context.Context) (*kafka.Kafka, error) {
	if a.kafkaClient == nil {
		// Создаем producer
		producerCfg := kafka.DefaultProducerConfig(a.cfg.KafkaBrokers)
		producer := kafka.NewProducer(producerCfg, &a.logger)

		// Создаем Kafka-клиент
		kafkaClient := kafka.New(producer, a.cfg.KafkaBrokers, a.cfg.KafkaGroupID, &a.logger)

		a.kafkaClient = kafkaClient

		a.logger.Info().Msg("kafka client created")
	}

	return a.kafkaClient, nil
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
	var errs []error

	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}

	if a.kafkaClient != nil {
		if err := a.kafkaClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close kafka client: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing app: %v", errs)
	}

	return nil
}
