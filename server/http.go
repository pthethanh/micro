package server

import (
	"net/http"
	"os"
	"path/filepath"
)

type (
	// HandlerOptions hold information of a HTTP handler options.
	HandlerOptions struct {
		// mandatory
		p string
		h http.Handler

		// options
		m            []string
		q            []string
		hdr          []string
		prefix       bool
		interceptors []HTTPInterceptor
	}

	// HTTPInterceptor is an interceptor/middleware func.
	HTTPInterceptor = func(http.Handler) http.Handler

	// spaHandler implements the http.Handler interface, so we can use it
	// to respond to HTTP requests. The path to the static directory and
	// path to the index file within that static directory are used to
	// serve the SPA in the given static directory.
	spaHandler struct {
		dir   string
		index string
	}
)

// NewHandlerOptions return new empty HTTP options.
func NewHandlerOptions() *HandlerOptions {
	return &HandlerOptions{}
}

// Prefix mark that the HTTP handler is a prefix handler.
func (r *HandlerOptions) Prefix() *HandlerOptions {
	r.prefix = true
	return r
}

// Methods adds a matcher for HTTP methods. It accepts a sequence of one or more methods to be matched.
func (r *HandlerOptions) Methods(methods ...string) *HandlerOptions {
	r.m = methods
	return r
}

// Queries adds a matcher for URL query values. It accepts a sequence of key/value pairs. Values may define variables.
func (r *HandlerOptions) Queries(queries ...string) *HandlerOptions {
	r.q = queries
	return r
}

// Headers adds a matcher for request header values. It accepts a sequence of key/value pairs to be matched.
func (r *HandlerOptions) Headers(headers ...string) *HandlerOptions {
	r.hdr = headers
	return r
}

// Interceptors adds interceptors into the handler.
func (r *HandlerOptions) Interceptors(interceptors ...HTTPInterceptor) *HandlerOptions {
	r.interceptors = interceptors
	return r
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// clean the paths
	path := filepath.Join(h.dir, filepath.Clean(r.URL.Path))

	// check whether a file exists at the given path
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.dir, h.index))
		return
	}
	// other errors, response with error
	if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// otherwise, serve the file
	http.ServeFile(w, r, path)
}
