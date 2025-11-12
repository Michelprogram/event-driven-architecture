package main

import (
	"eda-users/internal"
	"eda-users/internal/web"
	"log"
	"net/http"
)

const (
	PORT = ":3001"
)

func main() {

	kafka := internal.NewKafkaClient()
	defer kafka.Close()

	database, err := web.NewDatabase()

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Serveur up and running...")
	err = http.ListenAndServe(PORT, web.NewServer(database, kafka))

	if err != nil {
		log.Fatalln(err)
	}

}
