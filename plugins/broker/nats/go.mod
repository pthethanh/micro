module github.com/pthethanh/micro/plugins/broker/nats

go 1.16

require (
	github.com/nats-io/nats-server/v2 v2.6.4 // indirect
	github.com/nats-io/nats.go v1.13.1-0.20211018182449-f2416a8b1483
	github.com/pthethanh/micro v0.2.1
)

replace github.com/pthethanh/micro => ../../../../micro
