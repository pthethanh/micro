package broker

type (
	// PublishOptions is a configuration holder for publish options.
	PublishOptions struct {
	}

	// SubscribeOptions is a configuration holder for subscriptions.
	SubscribeOptions struct {
		// AutoAck defaults to true. When a handler returns
		// with a nil error the message is acked.
		AutoAck bool
		// Subscribers with the same queue name
		// will create a shared subscription where each
		// receives a subset of messages.
		Queue string
	}

	// PublishOption is a func for config publish options.
	PublishOption func(*PublishOptions)

	// SubscribeOption is a func for config subscription.
	SubscribeOption func(*SubscribeOptions)
)

// Queue sets the name of the queue to share messages on
func Queue(name string) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Queue = name
	}
}

// DisableAutoAck will disable auto ack of messages
// after they have been handled.
func DisableAutoAck() SubscribeOption {
	return func(o *SubscribeOptions) {
		o.AutoAck = false
	}
}

// Apply apply the options.
func (op *SubscribeOptions) Apply(opts ...SubscribeOption) {
	for _, f := range opts {
		f(op)
	}
}
