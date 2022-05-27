package broker_test

import (
	"testing"

	"github.com/pthethanh/micro/broker"
)

type (
	person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
)

func TestNewMessage(t *testing.T) {
	p := person{
		Name: "Jack",
		Age:  22,
	}
	m := broker.Must(broker.NewMessage(p, "application/json", "version", "1"))

	if err := m.UnmarshalBodyTo(&p); err != nil {
		t.Fatal(err)
	}
	if m.GetMessageType() != "broker_test.person" {
		t.Fatalf("got message-type=%s, want message-type=broker_test.person", m.GetMessageType())
	}
	if m.GetHeader()["version"] != "1" {
		t.Fatalf("got version=%s, want version=1", m.GetHeader()["version"])
	}
	// change name, test marshal to body
	p.Name = "James"
	if err := m.MarshalToBody(p); err != nil {
		t.Fatal(err)
	}
	if err := m.UnmarshalBodyTo(&p); err != nil {
		t.Fatal(err)
	}
	if p.Name != "James" {
		t.Fatalf("got name=%s, want name=James", p.Name)
	}
	if m.GetContentType() != "application/json" {
		t.Fatalf("got content-type=%s, want content-type=application/json", m.GetContentType())
	}
	// use codec name
	m = broker.Must(broker.NewMessage(p, "json"))
	if err := m.UnmarshalBodyTo(&p); err != nil {
		t.Fatal(err)
	}
	if p.Name != "James" {
		t.Fatalf("got name=%s, want name=James", p.Name)
	}
}
