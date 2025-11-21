package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Order struct {
    ProductID int `json:"productId"`
    Quantity  int `json:"quantity"`
    UserID    int `json:"userId"`
}

type Notification struct {
	Action string `json:"action"`
}

var ordersCollection *mongo.Collection

func main() {
    // Connexion MongoDB
    clientOpts := options.Client().ApplyURI("mongodb://admin_order:password_order@order-service-database:27017")
    client, err := mongo.Connect(context.Background(), clientOpts)
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

    kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
        Brokers:  []string{kafkaBroker},
        Topic:    "order-created",
        Balancer: &kafka.LeastBytes{},
    })
    defer kafkaWriter.Close()

    kafkaWriterNotification := kafka.NewWriter(kafka.WriterConfig{
        Brokers:  []string{kafkaBroker},
        Topic:    "notifications.central",
        Balancer: &kafka.LeastBytes{},
    })
    defer kafkaWriterNotification.Close()

    //Kakfa reader de commande, ajoute à la BDD
    //Une route des commandes pour les lister 
    

    // Router
    r := gin.Default()
    r.POST("/order", func(c *gin.Context) {
        var order Order
        if err := c.ShouldBindJSON(&order); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Sauvegarde MongoDB
        _, err := ordersCollection.InsertOne(context.Background(), order)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save order"})
            return
        }

        // Publier sur Kafka
        orderJson, err := json.Marshal(order)

        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal order"})
            return
        }

        if err = kafkaWriter.WriteMessages(context.Background(),
            kafka.Message{Value: orderJson},
        ); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish Kafka message"})
            return
        }



        notif := Notification{Action: fmt.Sprintf("New order created for ProductID %d", order.ProductID)}
        
        notifJson, err := json.Marshal(notif)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal notification"})
            return
        }

        kafkaWriterNotification.WriteMessages(context.Background(),
            kafka.Message{Value: notifJson},
        )

        c.JSON(http.StatusCreated, gin.H{"message": "Order created", "order": order})
    })

    
    r.Run(":3003")
}
