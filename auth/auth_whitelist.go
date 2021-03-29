package auth

import (
	"context"
	"regexp"
	"sync"
)

type (
	// WhiteListAuthenticator is a special authenticator that support ignoring a list
	// of methods/paths during the authentication process.
	WhiteListAuthenticator interface {
		Authenticator

		// IsWhiteListed tell the handler if a path should be ignored in the authentication process.
		// For gRPC request the path will be fullMethod, i.e... /helloworld.Greeter/SayHello.
		// For HTTP request the path will be URL.Path.
		IsWhiteListed(path string) bool
	}

	// WhiteListFunc is a function that used for matching whitelist item.
	WhiteListFunc = func(path string) bool

	// SimpleWhiteListAuthenticator is simple implementation of WhiteListAuthenticator.
	SimpleWhiteListAuthenticator struct {
		auth  Authenticator
		funcs []WhiteListFunc
		cache *sync.Map
	}
)

// NewWhiteListAuthenticator return a new WhiteListAuthenticator.
func NewWhiteListAuthenticator(auth Authenticator, funcs ...WhiteListFunc) *SimpleWhiteListAuthenticator {
	return &SimpleWhiteListAuthenticator{
		auth:  auth,
		funcs: funcs,
		cache: &sync.Map{},
	}
}

// Authenticate implements the Authenticator interface.
func (a *SimpleWhiteListAuthenticator) Authenticate(ctx context.Context) (context.Context, error) {
	return a.auth.Authenticate(ctx)
}

// IsWhiteListed implements WhitelistAuthenticator interface.
func (a *SimpleWhiteListAuthenticator) IsWhiteListed(path string) bool {
	// look at the cache first.
	if v, ok := a.cache.Load(path); ok && v.(bool) {
		return true
	}
	// check and update cache.
	for _, f := range a.funcs {
		if f(path) {
			a.cache.Store(path, true)
			return true
		}
	}
	return false
}

// WhiteListRegexp is white list function that ignore authentication process
// for a request if its path matchs one of the provided regular expressions.
// This function panic if the regular expressions failed to compile.
func WhiteListRegexp(patterns ...string) WhiteListFunc {
	regs := make([]*regexp.Regexp, 0)
	for _, p := range patterns {
		reg, err := regexp.Compile(p)
		if err != nil {
			panic(err)
		}
		regs = append(regs, reg)
	}
	return func(path string) bool {
		for _, reg := range regs {
			if reg.Match([]byte(path)) {
				return true
			}
		}
		return false
	}
}

// WhiteListInList is a simple white list func simply compare the path with the given list.
func WhiteListInList(wl ...string) WhiteListFunc {
	return func(path string) bool {
		for _, p := range wl {
			if p == path {
				return true
			}
		}
		return false
	}
}

// WhiteListNot is a helper function that simply reverse the inner white list func.
func WhiteListNot(f WhiteListFunc) WhiteListFunc {
	return func(path string) bool {
		return !f(path)
	}
}
