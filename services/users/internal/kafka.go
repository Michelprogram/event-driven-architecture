package internal

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

const (
	TOPIC          = "notifications.central"
	BROKER_ADDRESS = "kafka:29092"
)

type Notification struct {
	Action string `json:"action"`
}

type KafkaClient struct {
	writer *kafka.Writer
}

func NewKafkaClient() *KafkaClient {
	writer := &kafka.Writer{
		Addr:  kafka.TCP(BROKER_ADDRESS),
		Topic: TOPIC,
	}

	return &KafkaClient{
		writer,
	}
}

func (k KafkaClient) Send(message string) error {

	notification, err := json.Marshal(Notification{Action: message})
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte("notification"),
		Value: notification,
	}

	err = k.writer.WriteMessages(context.Background(), msg)
	if err != nil {
		return err
	}

	return nil

}

func (k KafkaClient) Close() {
	k.Close()
}
