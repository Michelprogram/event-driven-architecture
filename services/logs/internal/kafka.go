package internal

import (
	"context"
	"eda-logs/internal/types"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

const (
	TOPIC          = "logs"
	BROKER_ADDRESS = "kafka:29092"
	GROUP_ID       = "logs-group"
)

type KafkaClient struct {
	reader *kafka.Reader
	db     types.IDatabase
	logger chan<- string
}

func NewKafkaClient(logger chan<- string, db types.IDatabase) *KafkaClient {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{BROKER_ADDRESS},
		Topic:   TOPIC,
		GroupID: GROUP_ID,
	})

	return &KafkaClient{
		reader,
		db,
		logger,
	}
}

func (k KafkaClient) Read() error {

	for {
		m, err := k.reader.ReadMessage(context.Background())
		if err != nil {
			return err
		}

		var message types.Log
		if err := json.Unmarshal(m.Value, &message); err != nil {
			log.Printf("Error unmarshaling JSON: %v\n", err)
			continue
		}

		output := fmt.Sprintf("[Logs] Received %s", message)
		log.Println(output)
		k.logger <- output
		k.db.Save(message)
	}

}

func (k KafkaClient) Close() {
	k.Close()
}
