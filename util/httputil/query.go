package httputil

import "github.com/gorilla/schema"

const (
	defaultQueryTag = "query"
)

type (
	// QueryEncoder is a query param encoder.
	QueryEncoder = schema.Encoder
	// QueryDecoder is a query param decoder.
	QueryDecoder = schema.Decoder
)

// NewQueryDecoder return new query decoder using the given tag.
func NewQueryDecoder(tag string) *QueryDecoder {
	if tag == "" {
		tag = defaultQueryTag
	}
	dc := schema.NewDecoder()
	dc.SetAliasTag(tag)
	dc.IgnoreUnknownKeys(true)
	dc.ZeroEmpty(false)
	return dc
}

// NewQueryEncoder return new query encoder using the given tag.
func NewQueryEncoder(tag string) *QueryEncoder {
	if tag == "" {
		tag = defaultQueryTag
	}
	ec := schema.NewEncoder()
	ec.SetAliasTag(tag)
	return ec
}

// DecodeQuery decodes a map[string][]string to a struct.
//
// The first parameter must be a pointer to a struct.
//
// The second parameter is a map, typically url.Values from an HTTP request.
// Keys are "paths" in dotted notation to the struct fields and nested structs.
func DecodeQuery(dst interface{}, src map[string][]string) error {
	return NewQueryDecoder(defaultQueryTag).Decode(dst, src)
}

// EncodeQuery encodes a struct into map[string][]string.
//
// Intended for use with url.Values.
func EncodeQuery(src interface{}, dst map[string][]string) error {
	return NewQueryEncoder(defaultQueryTag).Encode(src, dst)
}
