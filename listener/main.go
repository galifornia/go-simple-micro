package main

import (
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// connect to rabbitmq
	rabbitMQ, err := connect()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer rabbitMQ.Close()
	log.Println("Connected to RabbitMQ")

	// listen for messages

	// create consumer
}

func connect() (*amqp.Connection, error) {
	counts := 0
	backOff := 1 * time.Second
	var connection *amqp.Connection
	for {
		c, err := amqp.Dial("amqp://guest:guest@localhost")
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
