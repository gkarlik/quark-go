package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/gkarlik/quark-go/broker"
	cb "github.com/gkarlik/quark-go/circuitbreaker"
	"github.com/gkarlik/quark-go/logger"
)

const (
	// Key defines Kafka message key.
	Key = "key"
	// Offset defines Kafka message offset.
	Offset = "offset"
	// Partition defines Kafka message partition.
	Partition = "partition"
	// Timestamp defines Kafka message timestamp.
	Timestamp = "timestamp"

	componentName = "KafkaBroker"
)

// MessageBroker represents message broker based on Kafka.
type MessageBroker struct {
	Consumer sarama.Consumer     // message consumer
	Producer sarama.SyncProducer // message producer
}

// NewMessageBroker creates instance of Kafka client message broker which is connected to provided addresses.
// Additional options passed as arguments are used to configure Kafka client and circuit breaker pattern to connect to Kafka instance.
// Panics if cannot create an instance (producer and/or consumer).
func NewMessageBroker(addrs []string, cfg *sarama.Config, opts ...cb.Option) *MessageBroker {
	consumer, err := new(cb.DefaultCircuitBreaker).Execute(func() (interface{}, error) {
		logger.Log().InfoWithFields(logger.Fields{
			"addrs":     addrs,
			"component": componentName,
		}, "Creating Kafka consumer")

		return sarama.NewConsumer(addrs, cfg)
	}, opts...)

	if err != nil {
		logger.Log().PanicWithFields(logger.Fields{
			"error":     err,
			"addrs":     addrs,
			"component": componentName,
		}, "Cannot create Kafka consumer")
	}

	producer, err := new(cb.DefaultCircuitBreaker).Execute(func() (interface{}, error) {
		logger.Log().InfoWithFields(logger.Fields{
			"addrs":     addrs,
			"component": componentName,
		}, "Creating Kafka producer")

		return sarama.NewSyncProducer(addrs, cfg)
	}, opts...)

	if err != nil {
		logger.Log().PanicWithFields(logger.Fields{
			"error":     err,
			"addrs":     addrs,
			"component": componentName,
		}, "Cannot create Kafka producer")
	}

	logger.Log().InfoWithFields(logger.Fields{
		"addrs":     addrs,
		"component": componentName,
	}, "Connected to Kafka broker")

	return &MessageBroker{
		Consumer: consumer.(sarama.Consumer),
		Producer: producer.(sarama.SyncProducer),
	}
}

// PublishMessage publishes message to Kafka broker.
func (b MessageBroker) PublishMessage(ctx context.Context, m broker.Message) error {
	logger.Log().InfoWithFields(logger.Fields{
		"message":   m,
		"component": componentName,
	}, "Publishing message")

	if b.Producer == nil {
		logger.Log().ErrorWithFields(logger.Fields{"component": componentName}, "Not connected to Kafka broker")

		return fmt.Errorf("[%s]: Not connected to Kafka server. Please check logs and network connection", componentName)
	}

	if m.Topic == "" {
		logger.Log().ErrorWithFields(logger.Fields{"component": componentName}, "Cannot publish message - message topic cannot be empty")

		return fmt.Errorf("[%s]: Cannot publish message - message topic cannot be empty", componentName)
	}

	body, err := json.Marshal(m.Value)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"message":   m,
			"component": componentName,
		}, "Cannot parse message body")

		return err
	}

	var key string
	if k, ok := m.Context[Key]; ok {
		key = k.(string)
	}

	msg := &sarama.ProducerMessage{Topic: m.Topic, Key: sarama.StringEncoder(key), Value: sarama.StringEncoder(body)}
	partition, offset, err := b.Producer.SendMessage(msg)

	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
			"topic":     m.Topic,
			"value":     m.Value,
			"key":       key,
		}, "Cannot publish message")

		return err
	}

	logger.Log().InfoWithFields(logger.Fields{
		"component": componentName,
		"key":       key,
		"topic":     m.Topic,
		"partition": partition,
		"offset":    offset,
		"value":     m.Value,
	}, "Message successfully published")

	return nil
}

// Subscribe subscribes to specified topic in Kafka broker.
// Using context (ctx) parameter it is possible to pass additional arguments such as kafka.Partition (int32) and kafka.Offset (int64).
func (b MessageBroker) Subscribe(ctx context.Context, topic string) (<-chan broker.Message, error) {
	logger.Log().InfoWithFields(logger.Fields{
		"topic":     topic,
		"component": componentName,
	}, "Subscribing to messages with topic")

	if b.Consumer == nil {
		logger.Log().ErrorWithFields(logger.Fields{"component": componentName}, "Not connected to Kafka broker")

		return nil, fmt.Errorf("[%s]: Not connected to Kafka server. Please check logs and network connection", componentName)
	}

	if topic == "" {
		logger.Log().ErrorWithFields(logger.Fields{"component": componentName}, "Cannot subscribe to messages - message topic cannot be empty")

		return nil, fmt.Errorf("[%s]: Cannot subscribe to messages - message topic cannot be empty", componentName)
	}

	var partition int32
	if p := ctx.Value(Partition); p != nil {
		partition = p.(int32)
	}

	offset := sarama.OffsetNewest
	if o := ctx.Value(Offset); o != nil {
		offset = o.(int64)
	}

	partitionConsumer, err := b.Consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
			"topic":     topic,
			"partition": partition,
			"offset":    offset,
		}, "Cannot consume message")
	}

	mgs := make(chan broker.Message)

	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				mgs <- broker.Message{
					Topic: msg.Topic,
					Value: string(msg.Value),
					Context: map[string]interface{}{
						Key:       msg.Topic,
						Offset:    msg.Offset,
						Partition: msg.Partition,
						Timestamp: msg.Timestamp,
					},
				}
			}
		}
	}()

	return mgs, nil
}

// Dispose closes Kafka client instance.
func (b *MessageBroker) Dispose() {
	logger.Log().InfoWithFields(logger.Fields{"component": componentName}, "Disposing message broker component")

	if b.Consumer != nil {
		b.Consumer.Close()
		b.Consumer = nil
	}

	if b.Producer != nil {
		b.Producer.Close()
		b.Producer = nil
	}
}
