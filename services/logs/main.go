package main

import (
	"eda-logs/internal"
	"log"
	"net/http"
)

const (
	PORT = ":3000"
)

func main() {

	logger := make(chan string, 1)

	client := internal.NewKafkaClient(logger)

	go func() { client.Read() }()
	defer client.Close()

	log.Println("Serveur up and running...")
	err := http.ListenAndServe(PORT, internal.NewWebsocketHandler(logger))

	if err != nil {
		log.Fatalln("Can't run websocket server: ", err)
	}
}
