package server

import (
	"net/textproto"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// DefaultHeaderMatcher return a ServerMuxOption that forward
// header keys request-id, api-key to GRPC Context.
func DefaultHeaderMatcher() runtime.ServeMuxOption {
	return HeaderMatcher([]string{"Request-Id", "Api-Key"})
}

// HeaderMatcher return a serveMuxOption for matcher header
// for passing a set of non IANA headers to GRPC context
// without a need to prefix them with Grpc-Metadata.
func HeaderMatcher(keys []string) runtime.ServeMuxOption {
	return runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
		for _, k := range keys {
			if k == canonicalKey || textproto.CanonicalMIMEHeaderKey(k) == canonicalKey {
				return k, true
			}
		}
		return runtime.DefaultHeaderMatcher(key)
	})
}
