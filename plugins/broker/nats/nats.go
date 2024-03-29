// Package nats provide a message broker using NATS.
package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pthethanh/micro/broker"
	"github.com/pthethanh/micro/encoding"
	"github.com/pthethanh/micro/health"
	"github.com/pthethanh/micro/log"
	"github.com/pthethanh/micro/util/syncutil"
)

type (
	// Nats is an implementation of broker.Broker using NATS.
	Nats struct {
		conn *nats.Conn
		opts []nats.Option
		log  log.Logger

		addrs string
		codec encoding.Codec
	}

	// Option is an optional configuration.
	Option func(*Nats)
)

var (
	_ broker.Broker  = (*Nats)(nil)
	_ health.Checker = (*Nats)(nil)
)

// New return a new NATs message broker.
// If address is not set, default address "nats:4222" will be used.
func New(opts ...Option) *Nats {
	n := &Nats{}
	// apply the options.
	for _, opt := range opts {
		opt(n)
	}
	if n.addrs == "" {
		n.addrs = defaultAddr
	}
	if n.log == nil {
		n.log = log.Root()
	}
	if n.codec == nil {
		n.codec = encoding.GetCodec(encoding.ContentTypeProto)
	}
	return n
}

// Open connect to target server.
func (n *Nats) Open(ctx context.Context) error {
	n.log.Infof("nats: connecting to %s", n.addrs)
	conn, err := nats.Connect(n.addrs, n.opts...)
	if err != nil {
		return err
	}
	n.conn = conn
	n.log.Context(ctx).Infof("nats: connected to %s successfully", n.addrs)
	return nil
}

// Publish implements broker.Broker interface.
func (n *Nats) Publish(ctx context.Context, topic string, m *broker.Message, opts ...broker.PublishOption) error {
	b, err := n.codec.Marshal(m)
	if err != nil {
		return err
	}
	return n.conn.Publish(topic, b)
}

// Subscribe implements broker.Broker interface.
func (n *Nats) Subscribe(ctx context.Context, topic string, h broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	op := &broker.SubscribeOptions{
		AutoAck: true,
	}
	op.Apply(opts...)
	msgHandler := func(msg *nats.Msg) {
		m := broker.Message{}
		if err := n.codec.Unmarshal(msg.Data, &m); err != nil {
			n.log.Context(ctx).Errorf("nats: subscribe: decode failed, err: %v", err)
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

// CheckHealth implements health.Checker.
func (n *Nats) CheckHealth(ctx context.Context) error {
	if !n.conn.IsConnected() {
		return fmt.Errorf("nats: server status=%d", n.conn.Status())
	}
	return nil
}

// Close flush in-flight messages and close the underlying connection.
func (n *Nats) Close(ctx context.Context) error {
	if err := syncutil.WaitCtx(ctx, 5*time.Second, func(ctx context.Context) {
		err := n.conn.FlushWithContext(ctx)
		if err != nil {
			n.log.Context(ctx).Errorf("nats: flush failed, err: ", err)
		}
		n.conn.Close()
	}); err != nil {
		n.log.Context(ctx).Errorf("nats: close, err: %v", err)
	}
	return nil
}
