package broker

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type (
	// Encoder is an interface for encoding/decoding used by broker.
	Encoder interface {
		Encode(v *Message) ([]byte, error)
		Decode(b []byte, v *Message) error
	}

	// JSONEncoder JSON format encoder.
	JSONEncoder struct{}
	// ProtoEncoder proto buffer encoder.
	ProtoEncoder struct{}
)

// Encode implements Encoder interface.
func (e JSONEncoder) Encode(v *Message) ([]byte, error) {
	return protojson.Marshal(v)
}

// Decode implements Encoder interface.
func (e JSONEncoder) Decode(b []byte, v *Message) error {
	return protojson.Unmarshal(b, v)
}

// Encode implements Encoder interface.
func (e ProtoEncoder) Encode(v *Message) ([]byte, error) {
	return proto.Marshal(v)
}

// Decode implements Encoder interface.
func (e ProtoEncoder) Decode(b []byte, v *Message) error {
	return proto.Unmarshal(b, v)
}
