package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

const (
	TOPIC          = "logs.central"
	BROKER_ADDRESS = "localhost:9092"
)

type Log struct {
	Message     string `json:"message"`
	ServiceName string `json:"service_name"`
}

func (l Log) String() string {
	return fmt.Sprintf("from %s: %s\n", l.ServiceName, l.Message)
}

type KafkaClient struct {
	reader *kafka.Reader
	logger chan<- string
}

func NewKafkaClient(logger chan<- string) *KafkaClient {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{BROKER_ADDRESS},
		Topic:   TOPIC,
	})

	return &KafkaClient{
		reader,
		logger,
	}
}

func (k KafkaClient) Read() error {

	for {
		m, err := k.reader.ReadMessage(context.Background())
		if err != nil {
			return err
		}

		var message Log
		if err := json.Unmarshal(m.Value, &message); err != nil {
			log.Printf("Error unmarshaling JSON: %v\n", err)
			continue
		}

		output := fmt.Sprintf("[Logs] Received Log: %s.\n", message)
		log.Println(output)
		k.logger <- output
	}

}

func (k KafkaClient) Close() {
	k.Close()
}
