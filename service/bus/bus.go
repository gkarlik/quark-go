package bus

import (
	"carlos/quark/service"
)

// Message represents structure which will be passed to message bus
type Message struct {
	Key   string
	Value interface{}
}

// ServiceBus represents pub/sub mechanism
type ServiceBus interface {
	PublishMessage(message Message) error
	Subscribe(key string) (<-chan Message, error)

	service.Disposer
}
