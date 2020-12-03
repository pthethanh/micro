package memory_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pthethanh/micro/broker"
	"github.com/pthethanh/micro/broker/memory"
)

func TestBroker(t *testing.T) {
	b := memory.New()
	defer b.Close(context.Background())
	type Person struct {
		Name string
		Age  int
	}
	topic := "test"
	if err := b.Connect(); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if err := b.HealthCheck()(context.TODO()); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	ch := make(chan broker.Event, 100)
	// sub without group
	sub, err := b.Subscribe(topic, func(msg broker.Event) error {
		if err := msg.Ack(); err != nil {
			t.Error(err)
		}
		ch <- msg
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Unsubscribe()
	if sub.Topic() != topic {
		t.Errorf("got topic=%s, want topic=%s", sub.Topic(), topic)
	}
	// sub on the queue q1
	subGroup1, err := b.Subscribe(topic, func(msg broker.Event) error {
		ch <- msg
		return nil
	}, broker.Queue("q1"))
	if err != nil {
		t.Fatal(err)
	}
	defer subGroup1.Unsubscribe()
	// sub with the same group as the previous one - queue q1
	subGroup2, err := b.Subscribe(topic, func(msg broker.Event) error {
		ch <- msg
		return nil
	}, broker.Queue("q1"))
	if err != nil {
		t.Fatal(err)
	}
	defer subGroup2.Unsubscribe()
	want := Person{
		Name: "jack",
		Age:  22,
	}
	// send n messages
	n := 2
	for i := 0; i < n; i++ {
		m := mustNewMessage(json.Marshal, want, map[string]string{"type": "person"})
		if err := b.Publish(topic, m); err != nil {
			t.Fatal(err)
		}
	}
	// send another message to a topic no one subscribe should not impact to the result.
	m := mustNewMessage(json.Marshal, want, map[string]string{"type": "person"})
	if err := b.Publish("other-topic", m); err != nil {
		t.Fatal(err)
	}
	close(ch)
	cnt := 0
	for e := range ch {
		cnt++
		if e.Topic() != topic {
			t.Fatalf("got topic=%s, want topic=test", e.Topic())
		}
		got := Person{}
		if err := json.Unmarshal(e.Message().Body, &got); err != nil {
			t.Fatalf("got body=%v, want body=%v", got, want)
		}
		if typ, ok := e.Message().Header["type"]; !ok || typ != "person" {
			t.Fatalf("got type=%s, want type=%s", typ, "person")
		}
	}
	// should received n*2 messages: sub get n, subGroup1 + subGroup2 = n
	if cnt != n*2 {
		t.Fatalf("got number of messages=%d, want number of messages=%d", cnt, n*2)
	}
}

func mustNewMessage(enc func(v interface{}) ([]byte, error), body interface{}, header map[string]string) *broker.Message {
	b, err := enc(body)
	if err != nil {
		panic(fmt.Sprintf("broker: new message, err: %v", err))
	}
	return &broker.Message{
		Header: header,
		Body:   b,
	}
}
