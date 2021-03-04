package httputil

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
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

// MustReadMockHandlersFromFile read mock specifications from JSON file.
// This method panics if any error.
func MustReadMockHandlersFromFile(path string) []MockHandler {
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
// NOTICE: Don't use this function for production since it's not optimized for production.
func Mock(handlers ...MockHandler) http.Handler {
	r := mux.NewRouter()
	for i := 0; i < len(handlers); i++ {
		handler := handlers[i]
		route := r.Path(handler.Path).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// serve file if body is pointing to an external file.
			if v, ok := handler.ResponseBody.(string); ok && strings.HasPrefix(v, filePrefix) {
				name := v[len(filePrefix):]
				http.ServeFile(w, r, name)
				return
			}
			for k, v := range handler.Header {
				w.Header().Set(k, v)
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

func hashit(v interface{}) string {
	if v == nil {
		return ""
	}
	h := func(v interface{}) string {
		h := md5.New()
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		h.Write(b)
		return base64.StdEncoding.EncodeToString(h.Sum(nil))
	}
	if rc, ok := v.(io.ReadCloser); ok {
		var val interface{}
		if err := json.NewDecoder(rc).Decode(&val); err != nil {
			return ""
		}
		return h(val)
	}
	return h(v)
}
