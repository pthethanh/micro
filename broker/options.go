package broker

type (
	PublishOptions struct {
	}

	SubscribeOptions struct {
		// AutoAck defaults to true. When a handler returns
		// with a nil error the message is acked.
		AutoAck bool
		// Subscribers with the same queue name
		// will create a shared subscription where each
		// receives a subset of messages.
		Queue string
	}

	PublishOption func(*PublishOptions)

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

func (op *SubscribeOptions) Apply(opts ...SubscribeOption) {
	for _, f := range opts {
		f(op)
	}
}
