// Code generated by protoc-gen-go. DO NOT EDIT.
// source: broker/broker.proto

package broker

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Message struct {
	Header               map[string]string `protobuf:"bytes,1,rep,name=header,proto3" json:"header,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Body                 []byte            `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}
func (*Message) Descriptor() ([]byte, []int) {
	return fileDescriptor_09a300fef54c4624, []int{0}
}

func (m *Message) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Message.Unmarshal(m, b)
}
func (m *Message) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Message.Marshal(b, m, deterministic)
}
func (m *Message) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Message.Merge(m, src)
}
func (m *Message) XXX_Size() int {
	return xxx_messageInfo_Message.Size(m)
}
func (m *Message) XXX_DiscardUnknown() {
	xxx_messageInfo_Message.DiscardUnknown(m)
}

var xxx_messageInfo_Message proto.InternalMessageInfo

func (m *Message) GetHeader() map[string]string {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *Message) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

func init() {
	proto.RegisterType((*Message)(nil), "broker.Message")
	proto.RegisterMapType((map[string]string)(nil), "broker.Message.HeaderEntry")
}

func init() { proto.RegisterFile("broker/broker.proto", fileDescriptor_09a300fef54c4624) }

var fileDescriptor_09a300fef54c4624 = []byte{
	// 154 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4e, 0x2a, 0xca, 0xcf,
	0x4e, 0x2d, 0xd2, 0x87, 0x50, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x42, 0x6c, 0x10, 0x9e, 0x52,
	0x2f, 0x23, 0x17, 0xbb, 0x6f, 0x6a, 0x71, 0x71, 0x62, 0x7a, 0xaa, 0x90, 0x31, 0x17, 0x5b, 0x46,
	0x6a, 0x62, 0x4a, 0x6a, 0x91, 0x04, 0xa3, 0x02, 0xb3, 0x06, 0xb7, 0x91, 0xb4, 0x1e, 0x54, 0x0b,
	0x54, 0x81, 0x9e, 0x07, 0x58, 0xd6, 0x35, 0xaf, 0xa4, 0xa8, 0x32, 0x08, 0xaa, 0x54, 0x48, 0x88,
	0x8b, 0x25, 0x29, 0x3f, 0xa5, 0x52, 0x82, 0x49, 0x81, 0x51, 0x83, 0x27, 0x08, 0xcc, 0x96, 0xb2,
	0xe4, 0xe2, 0x46, 0x52, 0x2a, 0x24, 0xc0, 0xc5, 0x9c, 0x9d, 0x5a, 0x29, 0xc1, 0xa8, 0xc0, 0xa8,
	0xc1, 0x19, 0x04, 0x62, 0x0a, 0x89, 0x70, 0xb1, 0x96, 0x25, 0xe6, 0x94, 0xa6, 0x82, 0x75, 0x71,
	0x06, 0x41, 0x38, 0x56, 0x4c, 0x16, 0x8c, 0x49, 0x6c, 0x60, 0xe7, 0x19, 0x03, 0x02, 0x00, 0x00,
	0xff, 0xff, 0xa1, 0x74, 0x03, 0x86, 0xb5, 0x00, 0x00, 0x00,
}
