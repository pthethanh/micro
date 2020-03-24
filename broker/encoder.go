package broker

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
)

type (
	Encoder interface {
		Encode(v *Message) ([]byte, error)
		Decode(b []byte, v *Message) error
	}

	JSONEncoder  struct{}
	ProtoEncoder struct{}
)

func (e JSONEncoder) Encode(v *Message) ([]byte, error) {
	return json.Marshal(v)
}

func (e JSONEncoder) Decode(b []byte, v *Message) error {
	return json.Unmarshal(b, v)
}

func (e ProtoEncoder) Encode(v *Message) ([]byte, error) {
	return proto.Marshal(v)
}

func (e ProtoEncoder) Decode(b []byte, v *Message) error {
	return proto.Unmarshal(b, v)
}
