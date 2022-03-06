module github.com/pthethanh/micro/plugins/broker/nats

go 1.16

require (
	github.com/nats-io/nats-server/v2 v2.7.3 // indirect
	github.com/nats-io/nats.go v1.13.1-0.20220121202836-972a071d373d
	github.com/pthethanh/micro v0.2.1
)

replace github.com/pthethanh/micro => ../../../../micro
