// Package broker defines standard interface for a message broker.
package broker

import (
	"context"

	"github.com/pthethanh/micro/health"
)

type (
	// Broker is an interface used for asynchronous messaging.
	Broker interface {
		// Connect establish connect to the target server.
		Connect() error
		Publish(topic string, m *Message, opts ...PublishOption) error
		Subscribe(topic string, h Handler, opts ...SubscribeOption) (Subscriber, error)
		HealthCheck() health.CheckFunc

		// Close flush all in-flight messages and close underlying connection.
		// Close allows a context to control the duration
		// of a flush/close call. This context should be non-nil.
		// If a deadline is not set, a default deadline of 5s will be applied.
		Close(context.Context) error
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
