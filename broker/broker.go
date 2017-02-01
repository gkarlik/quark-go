package broker

import (
	"github.com/gkarlik/quark-go/system"
)

// MessageContext represents context which is passed with a message.
type MessageContext map[string]string

// Message represents structure which will be passed to message broker.
type Message struct {
	Key     string         // message key
	Value   interface{}    // message value - typically JSON payload
	Context MessageContext // message context
}

// MessageBroker represents pub/sub mechanism.
type MessageBroker interface {
	PublishMessage(message Message) error
	Subscribe(key string) (<-chan Message, error)

	system.Disposer
}
