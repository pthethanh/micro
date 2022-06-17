// Package memory provides a message broker using memory.
package memory

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/pthethanh/micro/broker"
	"github.com/pthethanh/micro/health"
	"github.com/pthethanh/micro/util/syncutil"
)

type (
	// Broker is a memory message broker.
	Broker struct {
		subs   map[string][]*subscriber
		mu     *sync.RWMutex
		ch     chan func() error
		worker int
		buf    int
		wg     *sync.WaitGroup
		opened bool
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

	Option func(*Broker)
)

var (
	_ broker.Broker  = (*Broker)(nil)
	_ health.Checker = (*Broker)(nil)

	// ErrInvalidConnectionState indicate that the connection has not been opened properly.
	ErrInvalidConnectionState = errors.New("invalid connection state")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// New return new memory broker.
func New(opts ...Option) *Broker {
	br := &Broker{
		subs:   make(map[string][]*subscriber),
		mu:     &sync.RWMutex{},
		worker: 100,
		buf:    10_000,
		wg:     &sync.WaitGroup{},
	}
	for _, opt := range opts {
		opt(br)
	}
	br.ch = make(chan func() error, br.buf)
	return br
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
	wg := sync.WaitGroup{}
	wg.Add(br.worker)
	br.wg.Add(br.worker)
	for i := 0; i < br.worker; i++ {
		go func() {
			wg.Done()
			defer br.wg.Done()
			for h := range br.ch {
				_ = h()
			}
		}()
	}
	wg.Wait()
	br.opened = true
	return nil
}

// Publish implements broker.Broker interface.
func (br *Broker) Publish(ctx context.Context, topic string, m *broker.Message, opts ...broker.PublishOption) error {
	if !br.opened {
		return ErrInvalidConnectionState
	}
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
		sub := sub
		if sub.opts.Queue != "" {
			queueSubs[sub.opts.Queue] = append(queueSubs[sub.opts.Queue], sub)
			continue
		}
		// broad cast
		br.ch <- func() error { return sub.h(env) }
	}
	// queue subscribers, send to only 1 single random subscriber in the list.
	for _, queueSub := range queueSubs {
		queueSub := queueSub
		idx := rand.Intn(len(queueSub))
		br.ch <- func() error { return queueSub[idx].h(env) }
	}
	return nil
}

// Subscribe implements broker.Broker interface.
func (br *Broker) Subscribe(ctx context.Context, topic string, h broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	if !br.opened {
		return nil, ErrInvalidConnectionState
	}
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

// CheckHealth implements health.Checker interface.
func (br *Broker) CheckHealth(ctx context.Context) error {
	if !br.opened {
		return ErrInvalidConnectionState
	}
	return nil
}

// Close implements broker.Broker interface.
func (br *Broker) Close(ctx context.Context) error {
	br.opened = false
	close(br.ch)
	syncutil.WaitCtx(ctx, 10*time.Millisecond, func(ctx context.Context) {
		br.wg.Wait()
	})
	// unsubscribe all subscribers.
	for _, subs := range br.subs {
		for _, sub := range subs {
			sub.Unsubscribe()
		}
	}
	return nil
}
