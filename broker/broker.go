package broker

import (
	"github.com/gkarlik/quark-go/system"
)

// Message represents structure which will be passed to message bus
type Message struct {
	Key   string
	Value interface{}
}

// MessageBroker represents pub/sub mechanism
type MessageBroker interface {
	PublishMessage(message Message) error
	Subscribe(key string) (<-chan Message, error)

	system.Disposer
}
