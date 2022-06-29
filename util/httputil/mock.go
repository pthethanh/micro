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
	// MockRequest hold HTTP mocks request specification.
	MockRequest struct {
		// Path can be a fixed string or a pattern. See more at github.com/gorilla/mux.
		Path    string            `yaml:"path" json:"path"`
		Methods []string          `yaml:"methods" json:"methods"`
		Headers map[string]string `yaml:"headers" json:"headers"`
	}

	// MockResponse hold HTTP mocks response specification.
	MockResponse struct {
		Status  int               `yaml:"status" json:"status"`
		Headers map[string]string `yaml:"headers" json:"headers"`
		// Body can be a file using prefix file://path.
		Body interface{} `yaml:"body" json:"body"`
	}

	// MockHandler hold HTTP mock specification.
	MockHandler struct {
		Request  MockRequest  `yaml:"request" json:"request"`
		Response MockResponse `yaml:"response" json:"response"`
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
// Mock use sub-router definition hence it should be used with option server.HTTPPrefixHandler.
// NOTICE: Don't use this function for production since it's not optimized for performance.
func Mock(handlers ...MockHandler) http.Handler {
	r := mux.NewRouter()
	for i := 0; i < len(handlers); i++ {
		handler := handlers[i]
		route := r.Path(handler.Request.Path).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for k, v := range handler.Response.Headers {
				w.Header().Set(k, v)
			}
			// serve file if body is pointing to an external file.
			if v, ok := handler.Response.Body.(string); ok && strings.HasPrefix(v, filePrefix) {
				// it's user predefined path hence we don't call path.Clean here.
				name := v[len(filePrefix):]
				f, err := os.Open(name)
				if err != nil {
					WriteError(w, http.StatusNotFound, err)
					return
				}
				defer f.Close()
				w.WriteHeader(handler.Response.Status)
				io.Copy(w, f)
				return
			}
			WriteJSON(w, handler.Response.Status, handler.Response.Body)
		})
		if len(handler.Request.Headers) > 0 {
			headers := make([]string, 0)
			for k, v := range handler.Request.Headers {
				headers = append(headers, k, v)
			}
			route.Headers(headers...)
		}
		if len(handler.Request.Methods) > 0 {
			route.Methods(handler.Request.Methods...)
		}
	}
	return r
}
