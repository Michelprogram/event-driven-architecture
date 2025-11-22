package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Item struct {
	SKU string `json:"sku" bson:"sku"`
	Qty int    `json:"qty" bson:"qty"`
}

type Order struct {
	OrderID   string `json:"orderId" bson:"orderId"`
	Reserved  []Item `json:"reserved" bson:"reserved"`
	Timestamp string `json:"timestamp" bson:"timestamp"`
}

type Notification struct {
	Action string `json:"action"`
}

var ordersCollection *mongo.Collection

func main() {
	// Connexion MongoDB
	clientOpts := options.Client().ApplyURI("mongodb://user_app:strong_app_password@orders-service-database:27017/orders-db?authSource=orders-db")
	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	ordersCollection = client.Database("orders-db").Collection("orders")
	log.Println("Connected to MongoDB")

	// Kafka producer: lire le broker depuis l'env si défini, sinon fallback kafka:29092
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:29092"
	}

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   "order.central",
	})
	defer kafkaReader.Close()

	kafkaWriterNotification := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaBroker},
		Topic:   "notifications.central",
	})
	defer kafkaWriterNotification.Close()

	r := gin.Default()
	r.GET("/orders", func(c *gin.Context) {
		cur, err := ordersCollection.Find(context.Background(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders"})
			return
		}
		var orders []bson.M
		if err := cur.All(context.Background(), &orders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read orders"})
			return
		}
		c.JSON(http.StatusOK, orders)
	})

	go func() {
		for {
			message, err := kafkaReader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Error reading message: %v\n", err)
				continue
			}

			log.Println("Received message: ", string(message.Value))
			var order Order
			if err := json.Unmarshal(message.Value, &order); err != nil {
				log.Printf("Error unmarshaling JSON: %v\n", err)
				continue
			}

			time.Sleep(5 * time.Second)

			_, err = ordersCollection.InsertOne(context.Background(), order)
			if err != nil {
				log.Printf("Error saving order: %v\n", err)
				continue
			}

			notif := Notification{Action: fmt.Sprintf("New order created for ProductID %s", order.OrderID)}

			notifJson, err := json.Marshal(notif)
			if err != nil {
				log.Printf("Error marshalling notification: %v\n", err)
				continue
			}

			kafkaWriterNotification.WriteMessages(context.Background(),
				kafka.Message{Value: notifJson},
			)

		}
	}()

	r.Run(":3003")
}
