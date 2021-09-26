package json

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pthethanh/micro/encoding"
	"google.golang.org/protobuf/encoding/protojson"
)

type (
	codec struct {
		m *runtime.JSONPb
	}
)

const (
	Name = "json"
)

func init() {
	encoding.RegisterCodec(newCodec())
}

func newCodec() *codec {
	return &codec{
		m: &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:  true,
				UseEnumNumbers: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		},
	}
}

func (m *codec) Marshal(v interface{}) ([]byte, error) {
	return m.m.Marshal(v)
}

func (m *codec) Unmarshal(data []byte, v interface{}) error {
	return m.m.Unmarshal(data, v)
}

func (m *codec) Name() string {
	return Name
}
