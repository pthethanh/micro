//+build integration_test

package nats_test

import (
	"encoding/json"
	"testing"

	"github.com/pthethanh/micro/broker"
	"github.com/pthethanh/micro/broker/nats"
)

func TestBroker(t *testing.T) {
	b, err := nats.NewWithEncoder("nats://localhost:4223", broker.JSONEncoder{})
	if err != nil {
		t.Fatal(err)
	}
	type Person struct {
		Name string
		Age  int
	}
	ch := make(chan broker.Event, 1)
	sub, err := b.Subscribe("test", func(msg broker.Event) error {
		ch <- msg
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Unsubscribe()
	want := Person{
		Name: "jack",
		Age:  22,
	}
	m := broker.MustNewMessage(json.Marshal, want, map[string]string{"type": "person"})
	if err := b.Publish("test", m); err != nil {
		t.Fatal(err)
	}
	e := <-ch
	if e.Topic() != "test" {
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
