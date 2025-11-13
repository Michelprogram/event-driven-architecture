package main

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"os"
	"sync"

	"github.com/segmentio/kafka-go"
)

// Order defines the structure of our order event
type Order struct {
	OrderID   string `json:"order_id"`
	ItemID    string `json:"item_id"`
	UserEmail string `json:"user_email"`
	// Other fields omitted for brevity
}

// Item defines the structure of a single product item within an order
type Item struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

// OrderPlaced defines the structure of the "commande.placee" event
type OrderPlaced struct {
	OrderID string  `json:"orderId"`
	UserID  string  `json:"userId"`
	Items   []Item  `json:"items"`
	Total   float64 `json:"total"`
}

// Missing defines a product that could not be fully reserved due to insufficient stock
type Missing struct {
	SKU       string `json:"sku"`
	Required  int    `json:"required"`
	Available int    `json:"available"`
}

// StockReserved defines the structure of the "stock.reserve" event produced when stock reservation succeeds
type StockReserved struct {
	OrderID   string `json:"orderId"`
	Reserved  []Item `json:"reserved"`
	Timestamp string `json:"timestamp"`
}

// StockFailed defines the structure of the "stock.echec" event produced when stock reservation fails
type StockFailed struct {
	OrderID   string    `json:"orderId"`
	Reason    string    `json:"reason"`
	Missing   []Missing `json:"missing"`
	Timestamp string    `json:"timestamp"`
}

// In-memory stock simulation used for reservation checks
var (
	invMu sync.Mutex
	stock = map[string]int{
		"deck-basic": 10,
		"wheel-52mm": 8,
		"truck-139":  5,
	}
	seen = map[string]bool{} // commandes déjà traitées
)

// checkAndReserve verifies product availability and reserves stock if possible
func checkAndReserve(orderID string, items []Item) (ok bool, reserved []Item, missing []Missing) {
	invMu.Lock()
	defer invMu.Unlock()

	if seen[orderID] {
		return true, items, nil
	}

	for _, it := range items {
		avail := stock[it.SKU]
		if avail < it.Qty {
			missing = append(missing, Missing{
				SKU: it.SKU, Required: it.Qty, Available: avail,
			})
		}
	}

	if len(missing) > 0 {
		return false, nil, missing
	}

	for _, it := range items {
		stock[it.SKU] = stock[it.SKU] - it.Qty
		reserved = append(reserved, it)
	}
	seen[orderID] = true
	return true, reserved, nil
}


// env retrieves the value of an environment variable or returns a default value

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}


// logToCentral sends a structured log message to the central Kafka topic

func logToCentral(w *kafka.Writer, level, event string, payload any, key string) {
	msg := map[string]any{
		"service": "inventory",
		"level":   level,
		"event":   event,
		"payload": payload,
		"ts":      time.Now().UTC().Format(time.RFC3339),
	}
	b, _ := json.Marshal(msg)
	_ = w.WriteMessages(context.Background(),
		kafka.Message{Key: []byte(key), Value: b},
	)
}

// Configuration variables loaded from environment, with default values

var (
	brokerAddr = env("KAFKA_BROKER", "kafka:29092")
	topicIn    = env("TOPIC_ORDER_PLACED", "commande.placee") // peut rester "orders" au début
	topicOK    = env("TOPIC_STOCK_RESERVED", "stock.reserve")
	topicKO    = env("TOPIC_STOCK_FAILED", "stock.echec")
	topicLogs  = env("TOPIC_LOGS", "logs.central")
	groupID    = env("KAFKA_GROUP_ID", "inventory-group")
)

// main initializes Kafka connections and processes incoming "commande.placee" events
func main() {
	// Kafka consumer
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddr},
		Topic:    topicIn,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer r.Close()

	// Kafka producers
	wOK := &kafka.Writer{Addr: kafka.TCP(brokerAddr), Topic: topicOK, Balancer: &kafka.Hash{}}
	wKO := &kafka.Writer{Addr: kafka.TCP(brokerAddr), Topic: topicKO, Balancer: &kafka.Hash{}}
	wLog := &kafka.Writer{Addr: kafka.TCP(brokerAddr), Topic: topicLogs, Balancer: &kafka.Hash{}}
	defer wOK.Close()
	defer wKO.Close()
	defer wLog.Close()

	log.Println(" Inventory Service started... (listening on:", topicIn, ")")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v\n", err)
			continue
		}

		// Essayer de décoder commande.placee
		var evt OrderPlaced
		if err := json.Unmarshal(m.Value, &evt); err != nil || evt.OrderID == "" {
			// fallback: ancien format
			var old Order
			if err2 := json.Unmarshal(m.Value, &old); err2 == nil && old.OrderID != "" && old.ItemID != "" {
				evt = OrderPlaced{
					OrderID: old.OrderID,
					Items:   []Item{{SKU: old.ItemID, Qty: 1}},
				}
			} else {
				log.Printf("Unrecognized message: %s", string(m.Value))
				logToCentral(wLog, "error", "inventory.unmarshal", map[string]any{"raw": string(m.Value)}, "")
				continue
			}
		}

		ok, reserved, missing := checkAndReserve(evt.OrderID, evt.Items)
		if ok {
			out := StockReserved{
				OrderID:   evt.OrderID,
				Reserved:  reserved,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
			b, _ := json.Marshal(out)
			_ = wOK.WriteMessages(context.Background(),
				kafka.Message{Key: []byte(evt.OrderID), Value: b},
			)
			log.Printf(" stock.reserve envoyé pour commande %s", evt.OrderID)
			logToCentral(wLog, "info", "stock.reserve", out, evt.OrderID)
		} else {
			out := StockFailed{
				OrderID:   evt.OrderID,
				Reason:    "STOCK_INSUFFICIENT",
				Missing:   missing,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
			b, _ := json.Marshal(out)
			_ = wKO.WriteMessages(context.Background(),
				kafka.Message{Key: []byte(evt.OrderID), Value: b},
			)
			log.Printf(" stock.echec pour commande %s (articles manquants: %d)", evt.OrderID, len(missing))
			logToCentral(wLog, "error", "stock.echec", out, evt.OrderID)
		}
	}
}