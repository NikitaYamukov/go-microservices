package config

import (
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

// Config содержит конфигурацию приложения
type Config struct {
	ServiceName string `env:"SERVICE_NAME" required:"true" default:"account-service"`
	AppEnv      string `env:"APP_ENV" required:"true" default:"development"`
	Host        string `env:"GRPC_HOST" required:"true" default:"localhost"`
	Port        int    `env:"GRPC_PORT" required:"true" default:"50051"`
	LogLevel    string `env:"LOG_LEVEL" required:"true" default:"info"`

	DbDsn string `env:"DB_DSN" required:"true"`

	KafkaBrokers          []string `env:"KAFKA_BROKER_HOST" envSeparator:"," required:"true"`
	KafkaGroupID          string   `env:"KAFKA_CONSUMER_GROUP" required:"true"`
	KafkaTransactionTopic string   `env:"KAFKA_TRANSACTION_TOPIC" required:"true"`
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {

	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
