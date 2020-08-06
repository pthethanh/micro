package auth

import (
	"context"
	"regexp"
)

type (
	// WhiteListAuthenticator is a special authenticator that support ignoring a list
	// of methods that not required to be authenticated.
	WhiteListAuthenticator interface {
		Authenticator
		IsWhiteListed(fullMethod string) bool
	}

	// WhiteListMatchFunc is a function that used for matching whitelist item.
	// Some common match funcs (but not limitted to):
	// - strings.HasPrefix,
	// - strings.HasSuffix
	// - strings.Contains
	// - strings.ContainsAny
	// - strings.EqualFold
	// - auth.WhiteListMatchFuncRegexp
	// - auth.WhiteListMatchFuncExact
	WhiteListMatchFunc = func(fullMethod, whiteListPattern string) bool

	whitelistAuthenticator struct {
		auth    Authenticator
		wl      []string
		matches []WhiteListMatchFunc
	}
)

// NewWhiteListAuthenticator return a new WhiteListAuthenticator.
// If no matchFuncs is provided, default match func (exact match) will be used.
func NewWhiteListAuthenticator(auth Authenticator, whitelist []string, matchFuncs ...WhiteListMatchFunc) WhiteListAuthenticator {
	funcs := matchFuncs
	if len(funcs) == 0 {
		funcs = []WhiteListMatchFunc{WhiteListMatchFuncExact}
	}
	return &whitelistAuthenticator{
		auth:    auth,
		wl:      whitelist,
		matches: funcs,
	}
}

// Authenticate implements the Authenticator interface.
func (a *whitelistAuthenticator) Authenticate(ctx context.Context) (context.Context, error) {
	return a.auth.Authenticate(ctx)
}

// IsWhiteListed implements WhitelistAuthenticator interface.
func (a *whitelistAuthenticator) IsWhiteListed(fullMethod string) bool {
	for _, p := range a.wl {
		for _, f := range a.matches {
			if f(fullMethod, p) {
				return true
			}
		}
	}
	return false
}

// WhiteListMatchFuncRegexp is white list match func using regular expression.
func WhiteListMatchFuncRegexp(m string, p string) bool {
	if matched, err := regexp.MatchString(p, m); err == nil && matched {
		return true
	}
	return false
}

// WhiteListMatchFuncExact is a simple white list match func using string comparison.
func WhiteListMatchFuncExact(m string, p string) bool {
	return p == m
}
