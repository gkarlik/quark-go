package rabbitmq

import (
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/gkarlik/quark/service/bus"
	"github.com/streadway/amqp"
)

type messageBus struct {
	connection *amqp.Connection
}

// NewMessageBus creates instance of RabbotMQ Message Bus which is connected to specified address
func NewMessageBus(address string) *messageBus {
	log.WithField("address", address).Info("Connecting to RabbitMQ")

	conn, err := amqp.Dial(address)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"address": address,
		}).Error("Cannot connect to RabbitMQ")
	}
	return &messageBus{connection: conn}
}

// PublishMessage publishes message to RabbitMQ Message Bus
func (b messageBus) PublishMessage(m bus.Message) error {
	log.WithField("message", m).Info("Publishing message")

	if b.connection == nil {
		log.Error("Not connected to RabbitMQ")

		return errors.New("Not connected to RabbitMQ. Please check logs and network connection")
	}

	if m.Key == "" {
		log.Error("Message key cannot be empty")

		return errors.New("Message key cannot be empty")
	}

	ch, err := b.connection.Channel()
	if err != nil {
		log.Error("Cannot create channel")

		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		m.Key, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"queue": m.Key,
		}).Error("Cannot create queue")

		return err
	}

	body, err := json.Marshal(m.Value)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"message": m,
		}).Error("Cannot parse message body")
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"queue": q.Name,
		}).Error("Cannot publish message")
	}

	return nil
}

// Subscribe subscribes to specified routing key in RabbitMQ Message Bus
func (b messageBus) Subscribe(key string) (<-chan bus.Message, error) {
	log.WithField("key", key).Info("Subscribing to messages with key")

	if b.connection == nil {
		log.Error("Not connected to AMQP broker")

		return nil, errors.New("Not connected to AMQP broker. Please check logs and network connection")
	}

	if key == "" {
		log.Error("Key cannot be empty")

		return nil, errors.New("Key cannot be empty")
	}

	ch, err := b.connection.Channel()
	if err != nil {
		log.Error("Cannot create channel")

		return nil, err
	}

	q, err := ch.QueueDeclare(
		key,   // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"queue": key,
		}).Error("Cannot create queue")

		return nil, err
	}

	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"queue": q.Name,
		}).Error("Cannot consume message")
	}

	mgs := make(chan bus.Message)
	go func() {
		for msg := range messages {
			mgs <- bus.Message{
				Key:   q.Name,
				Value: msg.Body,
			}
		}
	}()

	return mgs, nil
}

// Close closes RabbitMQ Message Bus
func (b messageBus) Dispose() {
	if b.connection != nil {
		b.connection.Close()
	}
}
