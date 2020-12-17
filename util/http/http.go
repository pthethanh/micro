package http

import (
	"encoding/json"
	"fmt"
	"net/http"

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
// For writing error as text, use Write
func WriteError(w http.ResponseWriter, code int, err error) {
	WriteJSON(w, code, status.JSON(err))
}

// WriteJSON write status and JSON data to http ResponseWriter.
func WriteJSON(w http.ResponseWriter, code int, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	Write(w, "application/json", code, b)
}
