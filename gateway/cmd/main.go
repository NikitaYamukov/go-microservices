package main

import (
	"context"
	"log"

	"github.com/NikitaYamukov/go-microservices/gateway/internal/app"
	"github.com/NikitaYamukov/go-microservices/gateway/internal/config"
	"github.com/NikitaYamukov/go-microservices/gateway/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger := logger.New(cfg)

	application := app.New(&appLogger, cfg)
	if err := application.Run(context.Background()); err != nil {
		appLogger.Error().Msgf("Failed to run application: %v", err)
		return
	}
}
