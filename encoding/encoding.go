package encoding

import (
	"log"

	"github.com/pthethanh/micro/encoding/json"
	"google.golang.org/grpc/encoding"
)

const (
	ContentTypeJSON  = "json"
	ContentTypeProto = "proto"
)

type (
	Codec = encoding.Codec
)

func init() {
	// init only if not manual initalized yet.
	if encoding.GetCodec(ContentTypeJSON) == nil {
		RegisterCodec(json.New())
	}
}

// GetCodec gets a registered Codec by content-subtype, or nil if no Codec is
// registered for the content-subtype.
//
// The content-subtype is expected to be lowercase.
func GetCodec(subContentType string) Codec {
	return encoding.GetCodec(subContentType)
}

// RegisterCodec registers the provided Codec for use with all gRPC clients and
// servers.
//
// The Codec will be stored and looked up by result of its Name() method, which
// should match the content-subtype of the encoding handled by the Codec.  This
// is case-insensitive, and is stored and looked up as lowercase.  If the
// result of calling Name() is an empty string, RegisterCodec will panic. See
// Content-Type on
// https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md#requests for
// more details.
//
// NOTE: this function must only be called during initialization time (i.e. in
// an init() function), and is not thread-safe.  If multiple Codec are
// registered with the same name, the one registered last will take effect.
func RegisterCodec(c Codec) {
	encoding.RegisterCodec(c)
}

// MustMarshal panic if err is not nil.
func MustMarshal(data []byte, err error) []byte {
	if err != nil {
		log.Panic(err)
	}
	return data
}
