package broker

import (
	"reflect"
	"strings"

	"github.com/pthethanh/micro/encoding"
	"github.com/pthethanh/micro/status"

	// register default codecs
	_ "github.com/pthethanh/micro/encoding/json"
	_ "google.golang.org/grpc/encoding/proto"
)

const (
	// ContentType is header key for content type.
	ContentType = "content-type"
	// MessageType is header key for type of message's body.
	MessageType = "message-type"

	applicationJSON = "application/json"
)

// NewMessage create new message from the given information.
// Message type will be automatically retrieved.
// ContentType is standard content type like: application/json, application/proto.
// Or it can be codec name: json, proto,...
// But in both cases, the sub-content type or codec should be registered in advance
// via encoding.RegisterCodec so that it can be used for encoding/decoding the message body.
func NewMessage(message any, contentType string, headers ...string) (*Message, error) {
	if contentType == "" {
		contentType = applicationJSON
	}
	m := &Message{
		Header: map[string]string{
			ContentType: contentType,
			MessageType: GetMessageType(message),
		},
	}
	if len(headers)%2 == 1 {
		return nil, status.InvalidArgument("kv must be provided in pairs")
	}
	for i := 0; i < len(headers)/2; i++ {
		m.Header[headers[i]] = headers[i+1]
	}
	if err := m.MarshalToBody(message); err != nil {
		return nil, err
	}
	return m, nil
}

// Must panics if the given err is not nil.
func Must(m *Message, err error) *Message {
	if err != nil {
		panic(err)
	}
	return m
}

// GetMessageType return full type name of the given value without pointer indicator (*).
func GetMessageType(v any) string {
	return strings.TrimPrefix(reflect.TypeOf(v).String(), "*")
}

// UnmarshalBodyTo try to unmarshal the body of the message to the given pointer
// based on the content-type in the message's header.
// Use json codec for unmarshal if content-type is empty.
func (x *Message) UnmarshalBodyTo(ptr any) error {
	codec, err := x.getCodec()
	if err != nil {
		return err
	}
	return codec.Unmarshal(x.Body, ptr)
}

// MarshalToBody marshal the given value to the message body
// based on the registered content-type and the registered codec.
// Use json codec for marshal if content-type is empty.
func (x *Message) MarshalToBody(v any) error {
	codec, err := x.getCodec()
	if err != nil {
		return err
	}
	b, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	x.Body = b
	return nil
}

// GetContentType return content type configured in the message's header.
// Default to be json.
func (x *Message) GetContentType() string {
	if v, ok := x.Header[ContentType]; ok && v != "" {
		return v
	}
	return encoding.ContentTypeJSON
}

// getCodec return codec based on content type or codec configured in the message's header.
// Default to be json.
func (x *Message) getCodec() (encoding.Codec, error) {
	v, ok := x.Header[ContentType]
	if !ok || v == "" {
		return encoding.GetCodec(encoding.ContentTypeJSON), nil
	}
	cts := strings.Split(v, ";")
	ct := strings.Split(cts[0], "/")
	c := ct[0] // in case of codec name
	if len(ct) == 2 {
		c = ct[1] // in case of full content type.
	}
	codec := encoding.GetCodec(c)
	if codec == nil {
		return nil, status.Unimplemented("unregistered codec for content-type: %s", x.GetContentType())
	}
	return codec, nil
}

// GetMessageType return message type configured in the message's header.
// Otherwise return empty string.
func (x *Message) GetMessageType() string {
	return x.Header[MessageType]
}
