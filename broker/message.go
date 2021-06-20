package broker

import (
	"github.com/pthethanh/micro/encoding"
	"github.com/pthethanh/micro/status"
)

const (
	// ContentType is header key for content type.
	ContentType = "content-type"
)

// NewMessage create new message from the given information.
// If the given message is proto.Message and is not set,
// message type will be automatically retrieved.
func NewMessage(message interface{}, contentType string, kv ...string) (*Message, error) {
	m := &Message{
		Header: map[string]string{
			ContentType: contentType,
		},
	}
	if len(kv)%2 == 1 {
		return nil, status.InvalidArgument("kv must be provided in pairs")
	}
	for i := 0; i < len(kv)/2; i++ {
		m.Header[kv[i]] = kv[i+1]
	}
	if err := m.MarshalToBody(message); err != nil {
		return nil, err
	}
	return m, nil
}

// UnmarshalBodyTo try to unmarshal the body of the message to the given pointer
// based on the content-type in the message's header.
// Use json codec for unmarshal if content-type is empty.
func (x *Message) UnmarshalBodyTo(v interface{}) error {
	if m := encoding.GetCodec(x.GetContentType()); m != nil {
		return m.Unmarshal(x.Body, v)
	}
	return status.Unimplemented("unregistered codec for content-type: %s", x.GetContentType())
}

// MarshalToBody marshal the given value to the message body
// based on the registered content-type and the registered codec.
// Use json codec for marshal if content-type is empty.
func (x *Message) MarshalToBody(v interface{}) error {
	if m := encoding.GetCodec(x.GetContentType()); m != nil {
		b, err := m.Marshal(v)
		if err != nil {
			return err
		}
		x.Body = b
		return nil
	}
	return status.Unimplemented("unregistered codec for content-type: %s", x.GetContentType())
}

// GetContentType return content type configured in the message's header.
// Default to be json.
func (x *Message) GetContentType() string {
	if v, ok := x.Header[ContentType]; ok {
		return v
	}
	return encoding.ContentTypeJSON
}
