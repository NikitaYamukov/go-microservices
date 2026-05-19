package main

import (
	"context"
	"log"

	"github.com/NikitaYamukov/go-microservices/auth/internal/app"
	"github.com/NikitaYamukov/go-microservices/auth/internal/config"
	"github.com/NikitaYamukov/go-microservices/auth/internal/logger"
)

// @title Auth Service
// @version 1.0
// @description Auth Service
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
