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

	logger := make(chan string, 256)

	db, err := internal.NewDatabase()

	if err != nil {
		log.Fatalln("Can't connect to database: ", err)
	}

	client := internal.NewKafkaClient(db)

	go func() { client.Read(logger) }()
	defer client.Close()

	hub := internal.NewWebsocketHub(logger)
	go hub.Run()

	log.Println("Serveur up and running...")
	err = http.ListenAndServe(PORT, hub)

	if err != nil {
		log.Fatalln("Can't run websocket server: ", err)
	}
}
