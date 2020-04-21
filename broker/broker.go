// Package broker defines standard interface for a message broker.
package broker

import (
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
