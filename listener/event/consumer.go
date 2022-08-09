package event

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, s := range topics {
		err := ch.QueueBind(q.Name, s, "logs_topic", false, nil)
		if err != nil {
			return err
		}
	}

	messagesChannel, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messagesChannel {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)
			go handlePayload(payload)
		}
	}()

	log.Printf("Waiting for message on exchange on queue %s\n", q.Name)

	<-forever

	return nil
}

func handlePayload(payload Payload) error {
	switch payload.Name {
	case "log", "event":
		// log this thing
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
			return err
		}
	case "auth":
		// authenticate
	default:
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
			return err
		}

	}

	return nil
}

func logEvent(payload Payload) error {
	jsonData, _ := json.MarshalIndent(payload, "", "\t") // In production shoulld be Marshal wo Indent
	logServiceURL := "http://logger-service/log"         // !FIXME: read config from env

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}

	return nil
}
