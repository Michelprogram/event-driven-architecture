package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

const (
	LOG_TOPIC          = "logs.central"
	NOTIFICATION_TOPIC = "notifications.central"
	BROKER_ADDRESS     = "kafka:29092"
	GROUP_ID           = "notifications-group"
)

type Log struct {
	Message     string `json:"message"`
	ServiceName string `json:"service_name"`
}

type Notification struct {
	Action string `json:"action"`
}

type KafkaClient struct {
	reader *kafka.Reader
	writer *kafka.Writer
}

func NewKafkaClient() *KafkaClient {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{BROKER_ADDRESS},
		Topic:   NOTIFICATION_TOPIC,
		GroupID: GROUP_ID,
	})

	writer := &kafka.Writer{
		Addr:  kafka.TCP(BROKER_ADDRESS),
		Topic: LOG_TOPIC,
	}

	return &KafkaClient{
		reader,
		writer,
	}
}

func (k KafkaClient) Read() error {

	for {
		m, err := k.reader.ReadMessage(context.Background())
		if err != nil {
			return err
		}

		var message Notification
		if err := json.Unmarshal(m.Value, &message); err != nil {
			log.Printf("Error unmarshaling JSON: %v\n", err)
			continue
		}

		output := fmt.Sprintf("[Notification] Received Notification: %s", message.Action)
		log.Println(output)

		notification, err := json.Marshal(Log{Message: output, ServiceName: "notifications"})
		if err != nil {
			return err
		}

		msg := kafka.Message{
			Key:   []byte("notification"),
			Value: notification,
		}

		err = k.writer.WriteMessages(context.TODO(), msg)

		if err != nil {
			return err
		}
	}

}

func (k KafkaClient) Close() {
	k.Close()
}
