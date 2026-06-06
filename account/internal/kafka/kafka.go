package kafka

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

// Client представляет универсальный интерфейс для работы с Kafka
type Client interface {
	// Publish отправляет сообщение в указанный топик
	Publish(ctx context.Context, topic string, key string, data interface{}) error

	// Subscribe подписывается на топик и обрабатывает сообщения
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error

	// Close закрывает все соединения
	Close() error
}

// MessageHandler представляет функцию-обработчик сообщений
type MessageHandler func(ctx context.Context, topic string, key string, data []byte) error

// Kafka представляет основной клиент для работы с Kafka
type Kafka struct {
	producer *Producer
	brokers  []string
	groupID  string
	logger   *zerolog.Logger
}

// New создает новый экземпляр Kafka
func New(producer *Producer, brokers []string, groupID string, logger *zerolog.Logger) *Kafka {
	return &Kafka{
		producer: producer,
		brokers:  brokers,
		groupID:  groupID,
		logger:   logger,
	}
}

// Publish отправляет сообщение в указанный топик
func (k *Kafka) Publish(ctx context.Context, topic string, key string, data interface{}) error {
	return k.producer.SendJSONMessage(ctx, topic, key, data)
}

// Subscribe подписывается на топик и обрабатывает сообщения
func (k *Kafka) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	// Создаем уникальный GroupID для каждого топика
	uniqueGroupID := k.groupID + "-" + topic
	cfg := DefaultConsumerConfig(k.brokers, uniqueGroupID, []string{topic})
	consumer := NewConsumer(cfg, k.logger)

	// Запускаем consumer в отдельной горутине
	go func() {
		k.logger.Info().
			Str("topic", topic).
			Str("group_id", uniqueGroupID).
			Msg("subscribing to topic")

		messageHandler := ConsumerMessageHandler(func(ctx context.Context, msg kafka.Message) error {
			return handler(ctx, msg.Topic, string(msg.Key), msg.Value)
		})

		consumer.Start(ctx, messageHandler)
	}()

	return nil
}

// Close закрывает все соединения с Kafka
func (k *Kafka) Close() error {
	if k.producer != nil {
		if err := k.producer.Close(); err != nil {
			return fmt.Errorf("failed to close producer: %w", err)
		}
	}

	return nil
}
