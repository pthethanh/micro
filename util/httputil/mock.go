package httputil

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

const (
	filePrefix = "file://"
)

type (
	// MockHandler hold HTTP mock specification.
	MockHandler struct {
		Path   string            `yaml:"path" json:"path"`
		Method []string          `yaml:"method" json:"method"`
		Header map[string]string `yaml:"header" json:"header"`

		ResponseCode   int               `yaml:"response_code" json:"response_code"`
		ResponseHeader map[string]string `yaml:"response_header" json:"response_header"`
		// ResponseBody can be a file using file://path.
		ResponseBody interface{} `yaml:"response_body" json:"response_body"`
	}
)

// MustReadMockFromFile read mock specifications from JSON file.
// This function panics if any error.
func MustReadMockFromFile(path string) []MockHandler {
	rs := make([]MockHandler, 0)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, &rs); err != nil {
		panic(err)
	}
	return rs
}

// Mock provide a very simple way to mock a JSON HTTP handler base on path, method and header for testing.
// Mock use sub-router definition hence it should be used with option server.HTTPPrefixHandler instead of server.HTTPHandler.
// NOTICE: Don't use this function for production since it's not optimized for performance.
func Mock(handlers ...MockHandler) http.Handler {
	r := mux.NewRouter()
	for i := 0; i < len(handlers); i++ {
		handler := handlers[i]
		route := r.Path(handler.Path).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for k, v := range handler.ResponseHeader {
				w.Header().Set(k, v)
			}
			// serve file if body is pointing to an external file.
			if v, ok := handler.ResponseBody.(string); ok && strings.HasPrefix(v, filePrefix) {
				// it's user predefined path hence we don't call path.Clean here.
				name := v[len(filePrefix):]
				f, err := os.Open(name)
				if err != nil {
					WriteError(w, http.StatusNotFound, err)
					return
				}
				defer f.Close()
				w.WriteHeader(handler.ResponseCode)
				io.Copy(w, f)
				return
			}
			WriteJSON(w, handler.ResponseCode, handler.ResponseBody)
		})
		if len(handler.Header) > 0 {
			headers := make([]string, 0)
			for k, v := range handler.Header {
				headers = append(headers, k, v)
			}
			route.Headers(headers...)
		}
		if len(handler.Method) > 0 {
			route.Methods(handler.Method...)
		}
	}
	return r
}
