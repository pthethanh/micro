package json

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

type (
	Codec struct {
		m *runtime.JSONPb
	}
)

const (
	ContentType = "json"
)

func New() *Codec {
	return &Codec{
		m: &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
		},
	}
}

func (m *Codec) Marshal(v interface{}) ([]byte, error) {
	return m.m.Marshal(v)
}

func (m *Codec) Unmarshal(data []byte, v interface{}) error {
	return m.m.Unmarshal(data, v)
}

func (m *Codec) Name() string {
	return ContentType
}
