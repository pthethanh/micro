package broker

import (
	"fmt"

	"github.com/pthethanh/micro/health"
)

type (
	// Broker is an interface used for asynchronous messaging.
	Broker interface {
		Publish(topic string, m *Message, opts ...PublishOption) error
		Subscribe(topic string, h Handler, opts ...SubscribeOption) (Subscriber, error)
		HealthCheck() health.CheckFunc
	}

	// Handler is used to process messages via a subscription of a topic.
	// The handler is passed a publication interface which contains the
	// message and optional Ack method to acknowledge receipt of the message.
	Handler = func(Event) error

	// Event is given to a subscription handler for processing
	Event interface {
		Topic() string
		Message() *Message
		Ack() error
	}

	// Subscriber is a convenience return type for the Subscribe method
	Subscriber interface {
		Topic() string
		Unsubscribe() error
	}
)

// MustNewMessage return a new Message base on the given info.
// If any error, a panic will be thrown.
func MustNewMessage(enc func(v interface{}) ([]byte, error), body interface{}, header map[string]string) *Message {
	b, err := enc(body)
	if err != nil {
		panic(fmt.Sprintf("broker: new message, err: %v", err))
	}
	return &Message{
		Header: header,
		Body:   b,
	}
}
