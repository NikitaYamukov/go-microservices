package main

import (
	"log"

	_ "github.com/NikitaYamukov/go-microservices/docs"
	"github.com/NikitaYamukov/go-microservices/internal/config"
	"github.com/NikitaYamukov/go-microservices/internal/logger"
	"github.com/NikitaYamukov/go-microservices/internal/repository"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Account Service
// @version 1.0
// @description Account Service
// @host localhost:9000
// @BasePath /
func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем логгер
	logg := logger.New(cfg)

	// Подключаемся к базе данных
	db, err := gorm.Open(postgres.Open(cfg.DbDsn), &gorm.Config{})
	if err != nil {
		logg.Error().Msgf("Failed to connect to database: %v", err)
		return
	}

	logg.Info().Msg("database connected")

	// Инициализируем репозиторий
	repo := repository.NewRepository(db, logg)
	_ = repo

	// Gin router
	router := gin.Default()

	err = router.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		logg.Error().Msgf("Failed to set trusted proxies: %v", err)
		return
	}

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Test endpoint
	router.GET("/ping", PingExample)

	router.Run(":9000")
}

// PingExample godoc
// @Summary Проверка доступности сервиса
// @Description Возвращает pong
// @Tags health
// @Success 200 {string} string "pong"
// @Router /ping [get]
func PingExample(c *gin.Context) {
	c.String(200, "pong")
}
