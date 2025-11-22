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
	TOPIC          = "logs.central"
	BROKER_ADDRESS = "kafka:29092"
	GROUP_ID       = "logs-group"
)

type KafkaClient struct {
	reader *kafka.Reader
	db     types.IDatabase
}

func NewKafkaClient(db types.IDatabase) *KafkaClient {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{BROKER_ADDRESS},
		Topic:   TOPIC,
		GroupID: GROUP_ID,
	})

	return &KafkaClient{
		reader,
		db,
	}
}

func (k KafkaClient) Read(logger chan<- string) error {

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

		select {
		case logger <- output:
		default:
			// avoid blocking if no client yet; drop when buffer is full
			log.Println("logger channel full, dropping message")
		}

		if err := k.db.Save(message); err != nil {
			log.Printf("DB save error: %v", err)
		}
	}

}

func (k KafkaClient) Close() {
	k.Close()
}
