package main

import (
	"context"
	"log"

	"github.com/NikitaYamukov/go-microservices/internal/app"
	"github.com/NikitaYamukov/go-microservices/internal/config"
	"github.com/NikitaYamukov/go-microservices/internal/logger"
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

	application := app.New(appLogger, cfg)
	if err := application.Run(context.Background()); err != nil {
		appLogger.Error().Msgf("Failed to run application: %v", err)
		return
	}
}
