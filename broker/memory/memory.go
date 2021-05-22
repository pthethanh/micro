// Package memory provides a message broker using memory.
package memory

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/pthethanh/micro/broker"
	"github.com/pthethanh/micro/health"
)

type (
	// Broker is a memory message broker.
	Broker struct {
		subs map[string][]*subscriber
		mu   *sync.RWMutex
	}

	subscriber struct {
		id     string
		t      string
		h      broker.Handler
		opts   *broker.SubscribeOptions
		close  func()
		closed int32
	}

	event struct {
		t   string
		msg *broker.Message
	}
)

var (
	_ broker.Broker = &Broker{}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// New return new memory broker.
func New() *Broker {
	return &Broker{
		subs: make(map[string][]*subscriber),
		mu:   &sync.RWMutex{},
	}
}

func (env *event) Topic() string {
	return env.t
}

func (env *event) Message() *broker.Message {
	return env.msg
}

func (env *event) Ack() error {
	return nil
}

// Topic implements broker.Subscriber interface.
func (sub *subscriber) Topic() string {
	return sub.t
}

// Unsubscribe implements broker.Subscriber interface.
func (sub *subscriber) Unsubscribe() error {
	if atomic.AddInt32(&sub.closed, 1) > 1 {
		return nil
	}
	sub.close()
	return nil
}

// Open implements broker.Broker interface.
func (br *Broker) Open(ctx context.Context) error {
	// do nothing.
	return nil
}

// Publish implements broker.Broker interface.
func (br *Broker) Publish(ctx context.Context, topic string, m *broker.Message, opts ...broker.PublishOption) error {
	br.mu.RLock()
	subs := br.subs[topic]
	br.mu.RUnlock()
	// queue, list of sub
	queueSubs := make(map[string][]*subscriber)
	env := &event{
		t:   topic,
		msg: m,
	}
	for _, sub := range subs {
		if sub.opts.Queue != "" {
			queueSubs[sub.opts.Queue] = append(queueSubs[sub.opts.Queue], sub)
			continue
		}
		// broad cast
		sub.h(env)
	}
	// queue subscribers, send to only 1 single random subscriber in the list.
	for _, queueSub := range queueSubs {
		idx := rand.Intn(len(queueSub))
		queueSub[idx].h(env)
	}
	return nil
}

// Subscribe implements broker.Broker interface.
func (br *Broker) Subscribe(ctx context.Context, topic string, h broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	subOpts := &broker.SubscribeOptions{}
	subOpts.Apply(opts...)
	newSub := &subscriber{
		id:   uuid.New().String(),
		t:    topic,
		h:    h,
		opts: subOpts,
	}
	newSub.close = func() {
		br.mu.Lock()
		defer br.mu.Unlock()
		subs := br.subs[topic]
		// remove the sub
		newSubs := make([]*subscriber, 0)
		for _, sub := range subs {
			if newSub.id == sub.id {
				continue
			}
			newSubs = append(newSubs, sub)
		}
		br.subs[topic] = newSubs
	}
	br.mu.Lock()
	defer br.mu.Unlock()
	br.subs[topic] = append(br.subs[topic], newSub)
	return newSub, nil
}

// HealthCheck implements broker.Broker interface.
func (br *Broker) HealthCheck() health.CheckFunc {
	return func(ctx context.Context) error {
		// do nothing.
		return nil
	}
}

// Close implements broker.Broker interface.
func (br *Broker) Close(ctx context.Context) error {
	// unsubscribe all subsribers.
	for _, subs := range br.subs {
		for _, sub := range subs {
			sub.Unsubscribe()
		}
	}
	return nil
}
