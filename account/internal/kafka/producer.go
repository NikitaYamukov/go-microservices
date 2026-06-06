package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

// Producer представляет Kafka Producer
type Producer struct {
	writer *kafka.Writer
	logger *zerolog.Logger
}

// Message представляет сообщение для отправки в Kafka
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Headers   map[string]string
	Partition int
}

// NewProducer создает новый Kafka Producer
func NewProducer(cfg ProducerConfig, logger *zerolog.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,
		RequiredAcks: cfg.RequiredAcks,
		Async:        false,
	}

	return &Producer{
		writer: writer,
		logger: logger,
	}
}

// SendMessage отправляет сообщение в Kafka
func (p *Producer) SendMessage(ctx context.Context, msg Message) error {
	kafkaMsg := kafka.Message{
		Topic:     msg.Topic,
		Key:       msg.Key,
		Value:     msg.Value,
		Partition: msg.Partition,
		Time:      time.Now(),
	}

	// Добавляем заголовки
	for key, value := range msg.Headers {
		kafkaMsg.Headers = append(kafkaMsg.Headers, kafka.Header{
			Key:   key,
			Value: []byte(value),
		})
	}

	err := p.writer.WriteMessages(ctx, kafkaMsg)
	if err != nil {
		p.logger.Error().
			Err(err).
			Str("topic", msg.Topic).
			Str("key", string(msg.Key)).
			Msg("failed to send message to kafka")
		return fmt.Errorf("failed to send message to kafka: %w", err)
	}

	p.logger.Debug().
		Str("topic", msg.Topic).
		Str("key", string(msg.Key)).
		Msg("message sent to kafka successfully")

	return nil
}

// SendJSONMessage отправляет JSON-сообщение в Kafka
func (p *Producer) SendJSONMessage(ctx context.Context, topic string, key string, data interface{}) error {
	value, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message to JSON: %w", err)
	}

	msg := Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
		Headers: map[string]string{
			"content-type": "application/json",
		},
	}

	return p.SendMessage(ctx, msg)
}

// Close закрывает Producer
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}

	return nil
}
