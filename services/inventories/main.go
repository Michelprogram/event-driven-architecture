package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type Item struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

// Structure attendue pour l'événement 'payment.done' (qui déclenche la Tâche 1.5)
type OrderPlaced struct {
	OrderID string  `json:"orderId"`
	UserID  string  `json:"userId"`
	Items   []Item  `json:"items"`
	Total   float64 `json:"total"`
}

type Missing struct {
	SKU       string `json:"sku"`
	Required  int    `json:"required"`
	Available int    `json:"available"`
}

type StockReserved struct { // Événement de sortie: stock.reserve
	OrderID   string `json:"orderId"`
	Reserved  []Item `json:"reserved"`
	Timestamp string `json:"timestamp"`
}

type StockFailed struct { // Événement de sortie: stock.echec
	OrderID   string    `json:"orderId"`
	Reason    string    `json:"reason"`
	Missing   []Missing `json:"missing"`
	Timestamp string    `json:"timestamp"`
}

type Notification struct {
	Action string `json:"action"`
}

// Stock en mémoire et idempotence
var (
	invMu sync.Mutex
	stock = map[string]int{
		"pro-street":   10,
		"elite-deck":   8,
		"sunset-rider": 5,
		"park-master":  12,
	}
	seen = map[string]bool{}
)

// Vérifie et réserve le stock
func checkAndReserve(orderID string, items []Item) (ok bool, reserved []Item, missing []Missing) {
	invMu.Lock()
	defer invMu.Unlock()

	// Vérification d'idempotence: si la commande a déjà été traitée, on retourne true
	if seen[orderID] {
		log.Printf("Order %s already seen. Idempotence success.", orderID)
		return true, items, nil
	}

	// 1. Vérifie le stock
	for _, it := range items {
		avail := stock[it.SKU]
		log.Printf("Stock for %s: %d; Stock total: %v", it.SKU, avail, stock)
		if avail < it.Qty {
			missing = append(missing, Missing{
				SKU: it.SKU, Required: it.Qty, Available: avail,
			})
		}
	}

	// Si stock insuffisant, retourne l'échec
	if len(missing) > 0 {
		return false, nil, missing
	}

	// 2. Réserve le stock
	for _, it := range items {
		stock[it.SKU] -= it.Qty
		reserved = append(reserved, it)
	}

	// Marque comme traité
	seen[orderID] = true
	return true, reserved, nil
}

// Récupère une variable d'environnement ou valeur par défaut
func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	// Configuration
	brokerAddr := env("KAFKA_BROKER", "kafka:29092")

	// Tâche 1.5: Consomme l'événement 'payment.done'
	topicIn := env("TOPIC_INVENTORY_IN", "payment.done")

	// Tâche 1.5: Produit les événements 'stock.reserve' et 'stock.echec'
	topicOK := env("TOPIC_STOCK_RESERVED", "order.central")

	topicNotifications := env("TOPIC_NOTIFICATIONS", "notifications.central")
	groupID := env("KAFKA_GROUP_ID", "inventory-group")

	// Kafka consumer
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddr},
		Topic:    topicIn, // Écoute payment.done
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer r.Close()

	// Kafka producers
	wOK := &kafka.Writer{Addr: kafka.TCP(brokerAddr), Topic: topicOK, Balancer: &kafka.Hash{}}
	wNotifications := &kafka.Writer{Addr: kafka.TCP(brokerAddr), Topic: topicNotifications, Balancer: &kafka.Hash{}}
	defer func() {
		_ = wOK.Close()
		_ = wNotifications.Close()
	}()

	log.Printf("Inventory service started (listening on: %s)\n", topicIn)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v\n", err)
			continue
		}

		log.Println("Received message: ", string(m.Value))
		// Décodage du message (événement payment.done)
		var evt OrderPlaced
		if err := json.Unmarshal(m.Value, &evt); err != nil || evt.OrderID == "" {
			log.Printf("Unrecognized message: %s", string(m.Value))
			continue
		}

		log.Printf("Received payment.done for order %s. Processing stock check...", evt.OrderID)

		// Logique de la Tâche 1.5: Vérifie et réserve le stock
		ok, reserved, missing := checkAndReserve(evt.OrderID, evt.Items)

		if ok {
			// Stock réservé: Produit stock.reserve
			out := StockReserved{
				OrderID:   evt.OrderID,
				Reserved:  reserved,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
			b, _ := json.Marshal(out)
			if err := wOK.WriteMessages(context.Background(), kafka.Message{Key: []byte(evt.OrderID), Value: b}); err != nil {
				log.Printf("Error writing stock.reserve: %v", err)
			} else {
				log.Printf("stock.reserve sent for order %s", evt.OrderID)
			}

			notif := Notification{Action: fmt.Sprintf("Stock reserved for order %s", evt.OrderID)}
			notifJson, err := json.Marshal(notif)
			if err != nil {
				log.Printf("Error marshalling notification: %v\n", err)
				continue
			}
			wNotifications.WriteMessages(context.Background(), kafka.Message{Key: []byte(evt.OrderID), Value: notifJson})
		} else {
			// Stock insuffisant: Produit stock.echec
			notif := Notification{Action: fmt.Sprintf("Stock failed for order %s, missing items: %v", evt.OrderID, missing)}
			notifJson, err := json.Marshal(notif)
			if err != nil {
				log.Printf("Error marshalling notification: %v\n", err)
				continue
			}
			wNotifications.WriteMessages(context.Background(), kafka.Message{Key: []byte(evt.OrderID), Value: notifJson})
		}
	}
}
