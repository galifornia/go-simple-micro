package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const PORT = "80"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	rabbitMQ, err := connectToRabbit()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer rabbitMQ.Close()

	app := Config{
		Rabbit: rabbitMQ,
	}

	log.Printf("Starting Broker service in port %s\n", PORT)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connectToRabbit() (*amqp.Connection, error) {
	counts := 0
	backOff := 1 * time.Second
	var connection *amqp.Connection
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			log.Println("RabbitMQ is not yet ready")
			counts++
		} else {
			connection = c
			break
		}

		if counts == 5 {
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off")
		time.Sleep(backOff)
	}

	return connection, nil
}
