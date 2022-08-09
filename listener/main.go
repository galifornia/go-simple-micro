package main

import (
	"log"
	"math"
	"os"
	"time"

	"github.com/galifornia/go-micro-listener/event"

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

	// create consumer
	log.Println("Listening & consuming RabbitMQ messages")
	consumer, err := event.NewConsumer(rabbitMQ)
	if err != nil {
		panic(err)
	}

	// listen for messages
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println(err)
	}
}

func connect() (*amqp.Connection, error) {
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
