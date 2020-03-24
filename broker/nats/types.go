package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/pthethanh/micro/broker"
)

type (
	event struct {
		t string
		m *broker.Message
	}
	subscriber struct {
		t string
		s *nats.Subscription
	}
)

func (e *event) Topic() string {
	return e.t
}

func (e *event) Message() *broker.Message {
	return e.m
}

func (e *event) Ack() error {
	// nats does not support ack.
	return nil
}

func (s *subscriber) Topic() string {
	return s.t
}

func (s *subscriber) Unsubscribe() error {
	return s.s.Unsubscribe()
}
