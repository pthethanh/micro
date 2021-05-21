package httputil

import "github.com/gorilla/schema"

const (
	DefaultQueryTag = "query"
)

type (
	QueryEncoder = schema.Encoder
	QueryDecoder = schema.Decoder
)

// NewQueryDecoder return new query decoder using the given tag.
func NewQueryDecoder(tag string) *QueryDecoder {
	if tag == "" {
		tag = DefaultQueryTag
	}
	dc := schema.NewDecoder()
	dc.SetAliasTag(tag)
	return dc
}

// NewQueryEncoder return new query encoder using the given tag.
func NewQueryEncoder(tag string) *QueryEncoder {
	if tag == "" {
		tag = DefaultQueryTag
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
	dc := schema.NewDecoder()
	dc.SetAliasTag(DefaultQueryTag)
	dc.IgnoreUnknownKeys(true)
	dc.ZeroEmpty(false)
	return dc.Decode(dst, src)
}

// Encode encodes a struct into map[string][]string.
//
// Intended for use with url.Values.
func EncodeQuery(src interface{}, dst map[string][]string) error {
	ec := schema.NewEncoder()
	ec.SetAliasTag(DefaultQueryTag)
	return ec.Encode(src, dst)
}
