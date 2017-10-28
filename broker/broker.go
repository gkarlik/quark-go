package broker

import (
	"context"

	"github.com/gkarlik/quark-go/system"
)

// MessageContext represents context which is passed with a message.
type MessageContext map[string]interface{}

// Message represents structure which will be passed to message broker.
type Message struct {
	Topic   string         // message topic
	Value   interface{}    // message value - typically JSON payload
	Context MessageContext // message context
}

// MessageBroker represents pub/sub mechanism.
type MessageBroker interface {
	PublishMessage(ctx context.Context, message Message) error
	Subscribe(ctx context.Context, key string) (<-chan Message, error)

	system.Disposer
}
