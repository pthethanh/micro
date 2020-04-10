package nats

import "errors"

var (
	// ErrMissingEncoder report the encoder is missing.
	ErrMissingEncoder = errors.New("nats: missing encoder")
)
