// Package jwt implements authentication interfaces using JWT.
package jwt

import (
	"context"
	"strings"

	"github.com/pthethanh/micro/auth"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/metadata"
)

// Authenticator returns an AuthenticatorFunc that
// validates the provided JWT token in the :authorization header
// of the metadata.
func Authenticator(secret []byte) auth.AuthenticatorFunc {
	return func(ctx context.Context) (context.Context, error) {
		var claims Claims
		var newCtx context.Context
		if err := ParseFromMetadata(ctx, secret, &claims); err != nil {
			return newCtx, err
		}
		newCtx = NewContext(ctx, claims)
		return newCtx, nil
	}
}

// DecodeOnly returns an AuthenticatorFunc that ONLY try to decode JWT token in the
// authorization header of the metadata and attach the decoded user claims into the context.
// This authenticator does NOT return error in case the JWT is invalid
// or there is no authorization header in the metadata.
func DecodeOnly(secret []byte) auth.AuthenticatorFunc {
	return func(ctx context.Context) (context.Context, error) {
		var claims Claims
		if err := ParseFromMetadata(ctx, secret, &claims); err != nil {
			return ctx, nil
		}
		newCtx := NewContext(ctx, claims)
		return newCtx, nil
	}
}

// ParseFromMetadata fetches the JWT from the authorization metadata
// or in the grpcgateway-cookie located in the `Context`,
// validates the JWT and extracts the Claims.
func ParseFromMetadata(ctx context.Context, secret []byte, c jwt.Claims) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return auth.ErrMetadataMissing
	}
	// check from header
	slice, ok := md[auth.AuthorizationMD]
	if ok {
		if len(slice) > 1 {
			return auth.ErrMultipleAuthFound
		}
		return Parse(slice[0], secret, c)
	}
	// check from cookie
	for _, cookies := range md[auth.GrpcGWCookieMD] {
		for _, cookie := range strings.Split(cookies, ";") {
			slice := strings.Split(strings.TrimSpace(cookie), "=")
			if len(slice) == 2 && slice[0] == auth.AuthorizationMD {
				return Parse(slice[1], secret, c)
			}
		}
	}

	return auth.ErrAuthorizationMissing
}

// Parse and validate a JWT string.
func Parse(t string, s []byte, c jwt.Claims) error {
	_, err := jwt.ParseWithClaims(t, c, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, auth.ErrInvalidToken
		}
		return s, nil
	})
	if err != nil {
		return auth.ErrInvalidToken
	}
	return c.Valid()
}

// Encode encodes the jwt Claim to a JWT string.
func Encode(c jwt.Claims, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(secret)
}

// The context key
type claimsKey struct{}

// NewContext creates a new context with the claims attached.
func NewContext(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}

// FromContext fetches the claims attached to the ctx.
func FromContext(ctx context.Context) (c Claims, ok bool) {
	c, ok = ctx.Value(claimsKey{}).(Claims)
	return
}

// SubjectEquals checks if the JWT subject is equal to the provided
// subject in `sub`.
func SubjectEquals(ctx context.Context, s string) bool {
	if t, ok := FromContext(ctx); ok {
		return t.Subject == s
	}
	return false
}

// TokenString extracts the JWT toke as a string from `ctx`.
func TokenString(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	slice, ok := md[auth.AuthorizationMD]
	if !ok || len(slice) == 0 {
		return ""
	}
	if len(slice) > 1 {
		return ""
	}
	return slice[0]
}
