package kafka_test

import (
	"context"
	"github.com/Shopify/sarama"
	"testing"

	"encoding/json"
	"github.com/gkarlik/quark-go"
	"github.com/gkarlik/quark-go/broker"
	"github.com/gkarlik/quark-go/broker/kafka"
	cb "github.com/gkarlik/quark-go/circuitbreaker"
	"github.com/stretchr/testify/assert"
	"sync"
	"time"
)

type TestService struct {
	*quark.ServiceBase
}

type TestPayload struct {
	Text string `json:"text"`
}

func TestPublishSubscribe(t *testing.T) {
	topic := "TestTopic"
	text := "This is a test message"
	key, value := "TestKey", "TestValue"

	addr, _ := quark.GetHostAddress(1234)

	brokerAddr := "localhost:9092"

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Address(addr),
			quark.Broker(kafka.NewMessageBroker([]string{brokerAddr}, nil)),
		),
	}
	defer ts.Dispose()

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		partition := int32(0)
		offset := sarama.OffsetNewest

		ctx := context.WithValue(context.Background(), kafka.Partition, partition)
		ctx = context.WithValue(ctx, kafka.Offset, offset)

		messages, err := ts.Broker().Subscribe(ctx, topic)
		assert.NoError(t, err, "Subscribe returned an error")

		for msg := range messages {
			assert.Equal(t, topic, msg.Topic)

			var payload TestPayload

			err := json.Unmarshal([]byte(msg.Value.(string)), &payload)
			assert.NoError(t, err, "Unmarshal returned an error")

			assert.Equal(t, text, payload.Text)
			assert.Equal(t, int32(0), msg.Context[kafka.Partition])
			assert.NotEqual(t, sarama.OffsetNewest, msg.Context[kafka.Offset])
			wg.Done()
		}
	}()

	time.Sleep(1 * time.Second)

	go func() {
		ctx := broker.MessageContext{}
		ctx[key] = value
		ctx[kafka.Key] = "TestKey"

		m := broker.Message{
			Context: ctx,
			Topic:   topic,
			Value: &TestPayload{
				Text: text,
			},
		}

		err := ts.Broker().PublishMessage(context.Background(), m)
		assert.NoError(t, err, "Publish returned an error")

		wg.Done()
	}()

	wg.Wait()
}

func TestPublishSubscribeEmptyTopic(t *testing.T) {
	addr, _ := quark.GetHostAddress(1234)

	brokerAddr := "localhost:9092"

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Address(addr),
			quark.Broker(kafka.NewMessageBroker([]string{brokerAddr}, nil)),
		),
	}
	defer ts.Dispose()

	m := broker.Message{
		Value: "TestValue",
	}

	err := ts.Broker().PublishMessage(context.Background(), m)
	assert.Error(t, err, "Publish should return an error")

	_, err = ts.Broker().Subscribe(context.Background(), "")
	assert.Error(t, err, "Subscribe should return an error")
}

func TestBrokenConnection(t *testing.T) {
	addr, _ := quark.GetHostAddress(1234)

	brokerAddr := "localhost:9092"

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Address(addr),
			quark.Broker(kafka.NewMessageBroker([]string{brokerAddr}, nil)),
		),
	}
	defer ts.Dispose()

	m := broker.Message{
		Value: "TestValue",
	}

	// simulate broken connection
	b := ts.Broker().(*kafka.MessageBroker)
	b.Producer = nil

	err := ts.Broker().PublishMessage(context.Background(), m)
	assert.Error(t, err, "Publish should return an error")

	b.Consumer = nil

	_, err = ts.Broker().Subscribe(context.Background(), "")
	assert.Error(t, err, "Subscribe should return an error")
}

func TestBrokenNetworkConnection(t *testing.T) {
	addr, _ := quark.GetHostAddress(1234)
	brokerAddr := "wrong address"

	assert.Panics(t, func() {
		ts := &TestService{
			ServiceBase: quark.NewService(
				quark.Name("TestService"),
				quark.Version("1.0"),
				quark.Address(addr),
				quark.Broker(kafka.NewMessageBroker([]string{brokerAddr}, nil, cb.Retry(1), cb.Timeout(200*time.Microsecond))),
			),
		}
		defer ts.Dispose()
	})
}
