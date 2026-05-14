package main

import (
	"fmt"
	"log"
	"net"

	"github.com/NikitaYamukov/go-microservices/internal/config"
	"github.com/NikitaYamukov/go-microservices/internal/logger"
	"github.com/NikitaYamukov/go-microservices/internal/repository"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Account Service
// @version 1.0
// @description Account Service
// @host localhost:8080
// @BasePath /
func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем общий логгер с названием сервиса
	appLogger := logger.New(cfg)

	// Подключаемся к базе данных
	db, err := gorm.Open(postgres.Open(cfg.DbDsn), &gorm.Config{})
	if err != nil {
		appLogger.Error().Msgf("Failed to connect to database: %v", err)
		return
	}

	appLogger.Info().Msg("database connected")

	// Инициализируем репозиторий
	repo := repository.NewRepository(db, appLogger)
	_ = repo

	// Запускаем gRPC-сервер согласно контрактам
	listenAddr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		appLogger.Error().Msgf("Failed to listen on %s: %v", listenAddr, err)
		return
	}

	grpcServer := grpc.NewServer()

	// Регистрация health-сервиса и reflection для дебага
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthSrv)
	reflection.Register(grpcServer)

	appLogger.Info().Msg("service starting up")
	appLogger.Info().Msgf("gRPC server listening on %s", listenAddr)
	if err := grpcServer.Serve(lis); err != nil {
		appLogger.Error().Msgf("Failed to serve gRPC: %v", err)
		return
	}
}
