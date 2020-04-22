// Package nats provide a message broker using NATS.
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
	// Nats is an implementation of broker.Broker using NATS.
	Nats struct {
		conn *nats.Conn
		sub  chan *broker.Message
		opts []nats.Option

		addrs   string
		encoder broker.Encoder
	}

	// Option is an optional configuration.
	Option func(*Nats)
)

// New return a new NATs message broker.
// If address is not set, default address "nats:4222" will be used.
func New(opts ...Option) (*Nats, error) {
	n := &Nats{}
	// apply the options.
	for _, opt := range opts {
		opt(n)
	}
	if n.addrs == "" {
		n.addrs = defaultAddr
	}
	log.Debugf("nats: connecting to %s", n.addrs)
	conn, err := nats.Connect(n.addrs, n.opts...)
	if err != nil {
		return nil, err
	}
	log.Debugf("nats: connected to %s successfully", n.addrs)
	return &Nats{
		conn:    conn,
		encoder: n.encoder,
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
