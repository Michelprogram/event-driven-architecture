package main

import (
	"eda-notifications/internal"
	"log"
)

func main() {

	client := internal.NewKafkaClient()
	defer client.Close()

	log.Println("Starting notification service...")
	err := client.Read()

	if err != nil {
		log.Fatalf("Error reading messages: %v", err)
	}

}
