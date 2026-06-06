package kafka

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

// Consumer представляет Kafka Consumer
type Consumer struct {
	reader *kafka.Reader
	logger *zerolog.Logger
}

// GetConfig возвращает конфигурацию reader'а
func (c *Consumer) GetConfig() kafka.ReaderConfig {
	return c.reader.Config()
}

// ConsumerMessageHandler представляет функцию-обработчик сообщений для consumer
type ConsumerMessageHandler func(ctx context.Context, msg kafka.Message) error

// NewConsumer создает новый Kafka Consumer
func NewConsumer(cfg ConsumerConfig, logger *zerolog.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           cfg.Brokers,
		GroupID:           cfg.GroupID,
		Topic:             cfg.Topics[0], // Используем первый топик
		MinBytes:          cfg.MinBytes,
		MaxBytes:          cfg.MaxBytes,
		MaxWait:           cfg.MaxWait,
		ReadBatchTimeout:  cfg.ReadBatchTimeout,
		HeartbeatInterval: cfg.HeartbeatInterval,
		CommitInterval:    cfg.CommitInterval,
	})

	return &Consumer{
		reader: reader,
		logger: logger,
	}
}

// Start начинает потребление сообщений
func (c *Consumer) Start(ctx context.Context, handler ConsumerMessageHandler) error {
	c.logger.Info().
		Str("topic", c.reader.Config().Topic).
		Str("group_id", c.reader.Config().GroupID).
		Msg("starting kafka consumer")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("stopping kafka consumer")
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.Error().Err(err).Msg("failed to read message from kafka")
				time.Sleep(time.Second)
				continue
			}

			// Обрабатываем сообщение
			if err := handler(ctx, msg); err != nil {
				c.logger.Error().
					Err(err).
					Str("topic", msg.Topic).
					Int("partition", msg.Partition).
					Int64("offset", msg.Offset).
					Msg("failed to process message")
			} else {
				c.logger.Debug().
					Str("topic", msg.Topic).
					Int("partition", msg.Partition).
					Int64("offset", msg.Offset).
					Msg("message processed successfully")
			}
		}
	}
}

// Close закрывает Consumer
func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}

	return nil
}
