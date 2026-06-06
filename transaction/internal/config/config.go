package config

import (
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

// Config содержит конфигурацию приложения
type Config struct {
	ServiceName string `env:"SERVICE_NAME" json:"service_name" required:"true" default:"transaction-service"`
	AppEnv      string `env:"APP_ENV" json:"app_environment" required:"true" default:"development"`

	Host            string `env:"GRPC_HOST" json:"host" required:"true" default:"localhost"`
	Port            int    `env:"GRPC_PORT" json:"port" required:"true" default:"50054"`
	LogLevel        string `env:"LOG_LEVEL" json:"log_level" required:"true" default:"info"`
	DbDsn           string `env:"DB_DSN" json:"db_dsn" required:"true"`
	AccountGrpcHost string `env:"ACCOUNT_GRPC_HOST" json:"account_grpc_host" required:"true"`

	KafkaBrokers          []string `env:"KAFKA_BROKER_HOST" envSeparator:"," json:"kafka_brokers" required:"true"`
	KafkaGroupID          string   `env:"KAFKA_CONSUMER_GROUP" json:"kafka_consumer_group" required:"true"`
	KafkaTransactionTopic string   `env:"KAFKA_TRANSACTION_TOPIC" json:"kafka_transaction_topic" required:"true"`
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
