// Package broker defines standard interface for a message broker.
package broker

import (
	"context"

	"github.com/pthethanh/micro/health"
)

type (
	// Broker is an interface used for asynchronous messaging.
	Broker interface {
		// Open establish connection to the target server.
		Open(ctx context.Context) error
		// Publish publish the message to the target topic.
		Publish(ctx context.Context, topic string, m *Message, opts ...PublishOption) error
		// Subscribe subscribe to the topic to consume messages.
		Subscribe(ctx context.Context, topic string, h Handler, opts ...SubscribeOption) (Subscriber, error)
		// HealthCheck return health check function for checking health.
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
