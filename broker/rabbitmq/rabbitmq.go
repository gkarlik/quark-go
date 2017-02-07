package rabbitmq

import (
	"encoding/json"
	"fmt"

	"github.com/gkarlik/quark-go/broker"
	cb "github.com/gkarlik/quark-go/circuitbreaker"
	"github.com/gkarlik/quark-go/logger"
	"github.com/streadway/amqp"
)

const componentName = "RabbitMQBroker"

// MessageBroker represents message broker based on RabbitMQ
type MessageBroker struct {
	Connection *amqp.Connection // amqp connection
}

// NewMessageBroker creates instance of RabbitMQ message broker which is connected on provided address.
// Additional options passed as arguments are used to configure circuit breaker pattern to connect to RabbitMQ instance.
// Panics if cannot create an instance.
func NewMessageBroker(address string, opts ...cb.Option) *MessageBroker {
	conn, err := new(cb.DefaultCircuitBreaker).Execute(func() (interface{}, error) {
		logger.Log().InfoWithFields(logger.Fields{
			"address":   address,
			"component": componentName,
		}, "Connecting to RabbitMQ server")
		return amqp.Dial(address)
	}, opts...)

	if err != nil {
		logger.Log().PanicWithFields(logger.Fields{
			"error":     err,
			"address":   address,
			"component": componentName,
		}, "Cannot connect to RabbitMQ server")
	}
	logger.Log().InfoWithFields(logger.Fields{
		"address":   address,
		"component": componentName,
	}, "Connected to RabbitMQ server")

	return &MessageBroker{Connection: conn.(*amqp.Connection)}
}

// PublishMessage publishes message to RabbitMQ instance.
func (b MessageBroker) PublishMessage(m broker.Message) error {
	logger.Log().InfoWithFields(logger.Fields{
		"message":   m,
		"component": componentName,
	}, "Publishing message")

	if b.Connection == nil {
		logger.Log().ErrorWithFields(logger.Fields{"component": componentName}, "Not connected to RabbitMQ instance")

		return fmt.Errorf("[%s]: Not connected to RabbitMQ server. Please check logs and network connection", componentName)
	}

	if m.Key == "" {
		logger.Log().ErrorWithFields(logger.Fields{"component": componentName}, "Cannot publish message - message key cannot be empty")

		return fmt.Errorf("[%s]: Cannot publish message - message key cannot be empty", componentName)
	}

	ch, err := b.Connection.Channel()
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
		}, "Cannot create channel")

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
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"queue":     m.Key,
			"component": componentName,
		}, "Cannot create queue")

		return err
	}

	body, err := json.Marshal(m.Value)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"message":   m,
			"component": componentName,
		}, "Cannot parse message body")
	}

	// fill message headers with context
	headers := amqp.Table{}
	for k, v := range m.Context {
		headers[k] = v
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers:     headers,
		})

	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"queue":     q.Name,
			"component": componentName,
		}, "Cannot publish message")
	}

	return nil
}

// Subscribe subscribes to specified routing key in RabbitMQ instance.
func (b MessageBroker) Subscribe(key string) (<-chan broker.Message, error) {
	logger.Log().InfoWithFields(logger.Fields{
		"key":       key,
		"component": componentName,
	}, "Subscribing to messages with key")

	if b.Connection == nil {
		logger.Log().ErrorWithFields(logger.Fields{"component": componentName}, "Not connected to RabbitMQ server")

		return nil, fmt.Errorf("[%s]: Not connected to RabbitMQ server. Please check logs and network connection", componentName)
	}

	if key == "" {
		logger.Log().ErrorWithFields(logger.Fields{"component": componentName}, "Cannot subscribe to messages - Key cannot be empty")

		return nil, fmt.Errorf("[%s]: Cannot subscribe to messages - Key cannot be empty", componentName)
	}

	ch, err := b.Connection.Channel()
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
		}, "Cannot create channel")

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
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"queue":     key,
			"component": componentName,
		}, "Cannot create queue")

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
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"queue":     q.Name,
			"component": componentName,
		}, "Cannot consume message")
	}

	mgs := make(chan broker.Message)
	go func() {
		for msg := range messages {
			// create message context from headers
			context := broker.MessageContext{}
			for k, v := range msg.Headers {
				context[k] = v.(string)
			}

			mgs <- broker.Message{
				Key:     q.Name,
				Value:   msg.Body,
				Context: context,
			}
		}
	}()

	return mgs, nil
}

// Dispose closes RabbitMQ connection.
func (b *MessageBroker) Dispose() {
	logger.Log().InfoWithFields(logger.Fields{"component": componentName}, "Disposing message broker component")

	if b.Connection != nil {
		b.Connection.Close()
		b.Connection = nil
	}
}
