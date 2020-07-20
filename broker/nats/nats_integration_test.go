//+build integration_test

package nats_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pthethanh/micro/broker"
	"github.com/pthethanh/micro/broker/nats"
)

func TestBroker(t *testing.T) {
	b := nats.New(nats.Address("nats://localhost:4222"), nats.Encoder(broker.JSONEncoder{}))
	if err := b.Connect(); err != nil {
		t.Fatal(err)
	}
	defer b.Close(context.Background())
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
	m := mustNewMessage(json.Marshal, want, map[string]string{"type": "person"})
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
