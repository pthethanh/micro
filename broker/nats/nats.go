package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/pthethanh/micro/broker"
	"github.com/pthethanh/micro/health"
	"github.com/pthethanh/micro/log"
)

type (
	// Nats is an implementation of broker.Broker.
	Nats struct {
		conn    *nats.Conn
		encoder broker.Encoder
		sub     chan *broker.Message
	}
)

// New Nats with a configuration and additional options if needed.
func New(conf Config, additionalOpts ...nats.Option) (*Nats, error) {
	opts := make([]nats.Option, 0)
	opts = append(opts, nats.Timeout(conf.Timeout))
	if conf.Username != "" {
		opts = append(opts, nats.UserInfo(conf.Username, conf.Password))
	}
	opts = append(opts, additionalOpts...)
	return NewWithEncoder(conf.Addrs, conf.GetEncoder(), opts...)
}

// NewWithEncoder return a new NATS client with the given encoder.
func NewWithEncoder(addrs string, enc broker.Encoder, opts ...nats.Option) (*Nats, error) {
	conn, err := nats.Connect(addrs, opts...)
	if err != nil {
		return nil, err
	}
	return &Nats{
		conn:    conn,
		encoder: enc,
	}, nil
}

// Publish implements broker.Broker interface.
func (n *Nats) Publish(topic string, m *broker.Message, opts ...broker.PublishOption) error {
	if n.encoder == nil {
		return ErrMissingEncoder
	}
	b, err := n.encoder.Encode(m)
	if err != nil {
		return err
	}
	return n.conn.Publish(topic, b)
}

// Subscribe implements broker.Broker interface.
func (n *Nats) Subscribe(topic string, h broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	if n.encoder == nil {
		return nil, ErrMissingEncoder
	}
	op := &broker.SubscribeOptions{
		AutoAck: true,
	}
	op.Apply(opts...)
	msgHandler := func(msg *nats.Msg) {
		m := broker.Message{}
		if err := n.encoder.Decode(msg.Data, &m); err != nil {
			log.Errorf("nats: subscribe: decode failed, err: %v", err)
			return
		}
		h(&event{
			t: topic,
			m: &m,
		})
	}
	if op.Queue != "" {
		sub, err := n.conn.QueueSubscribe(topic, op.Queue, msgHandler)
		if err != nil {
			return nil, err
		}
		return &subscriber{
			t: topic,
			s: sub,
		}, nil
	}
	sub, err := n.conn.Subscribe(topic, msgHandler)
	if err != nil {
		return nil, err
	}
	return &subscriber{
		t: topic,
		s: sub,
	}, nil
}

// HealthCheck return a health check func.
func (n *Nats) HealthCheck() health.CheckFunc {
	return health.CheckFunc(func(context.Context) error {
		if !n.conn.IsConnected() {
			return fmt.Errorf("server status=%d", n.conn.Status())
		}
		return nil
	})
}
