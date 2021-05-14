package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/pthethanh/micro/status"
)

// Write write the status code and body on a http ResponseWriter
func Write(w http.ResponseWriter, contentType string, code int, body []byte) {
	w.Header().Set("Content-Length", fmt.Sprintf("%v", len(body)))
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	w.Write(body)
}

// WriteError write the status code and the error in JSON format on a http ResponseWriter.
// For writing error as plain text or other formats, use Write.
func WriteError(w http.ResponseWriter, code int, err error) {
	Write(w, "application/json", code, status.JSON(err))
}

// WriteJSON write status and JSON data to http ResponseWriter.
func WriteJSON(w http.ResponseWriter, code int, data interface{}) {
	if err, ok := data.(error); ok {
		WriteError(w, code, err)
		return
	}
	b, err := json.Marshal(data)
	if err != nil {
		WriteError(w, code, status.Internal("http: write json, err: %v", err))
		return
	}
	Write(w, "application/json", code, b)
}

// DecodeQuery decodes a map[string][]string to a struct.
//
// The first parameter must be a pointer to a struct.
//
// The second parameter is a map, typically url.Values from an HTTP request.
// Keys are "paths" in dotted notation to the struct fields and nested structs.
func DecodeQuery(dst interface{}, src map[string][]string) error {
	return schema.NewDecoder().Decode(dst, src)
}

// Encode encodes a struct into map[string][]string.
//
// Intended for use with url.Values.
func EncodeQuery(src interface{}, dst map[string][]string) error {
	return schema.NewEncoder().Encode(src, dst)
}
