package jwt

import (
	"context"
	"reflect"
	"testing"

	"github.com/pthethanh/micro/auth"
	"google.golang.org/grpc/metadata"
)

func TestContainScopes(t *testing.T) {
	claims := Claims{
		Scope: "foo bar foobar",
	}
	tt := []struct {
		scopes []string
		result bool
	}{
		{[]string{"foo"}, true},
		{[]string{"foo", "bar"}, true},
		{[]string{"foo", "bar", "foobar"}, true},
		{[]string{"foobar"}, true},
		{[]string{}, true},
		{[]string{""}, false},
		{[]string{"missing"}, false},
		{[]string{"foobar", "barfoo"}, false},
		{[]string{"foo", "bar", "foobar", ""}, false},
	}
	for _, tc := range tt {
		if got := claims.ContainScopes(tc.scopes...); got != tc.result {
			t.Errorf("claims.ContainScopes(%q) = %t; want %t", tc.scopes, got, tc.result)
		}
	}
}

func TestAuthenticator(t *testing.T) {
	secret := []byte("very-secret-secret")
	fn := Authenticator(secret)
	claims := Claims{
		Scope: "foo bar foobar",
	}
	token, err := Encode(claims, secret)
	if err != nil {
		t.Fatalf("Encode(%+v, %q) failed with: %v", claims, secret, err)
	}
	md := metadata.New(map[string]string{"authorization": token})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	newCtx, err := fn(ctx)
	if err != nil {
		t.Errorf("Authentication failed with: %v", err)
	}
	newClaims, ok := FromContext(newCtx)
	if !ok {
		t.Error("Claims missing from context")
	}
	if !reflect.DeepEqual(claims, newClaims) {
		t.Errorf("Claims not equal: %v vs %v", claims, newClaims)
	}
}

func TestParseFromMetadata(t *testing.T) {
	secret := []byte("very-secret-secret")
	token, err := Encode(Claims{}, secret)
	if err != nil {
		t.Fatalf("Encode returned an error: %v", err)
	}
	tt := []struct {
		meta metadata.MD
		err  error
	}{
		{nil, auth.ErrMetadataMissing},
		{metadata.Pairs("key", "value"), auth.ErrAuthorizationMissing},
		{metadata.Pairs("authorization", "key", "authorization", "newkey"), auth.ErrMultipleAuthFound},
		{metadata.Pairs("authorization", "token"), auth.ErrInvalidToken},
		{metadata.Pairs("authorization", token), nil},
	}
	for _, tc := range tt {
		var c Claims
		ctx := context.Background()
		if tc.meta != nil {
			ctx = metadata.NewIncomingContext(ctx, tc.meta)
		}
		if err := ParseFromMetadata(ctx, secret, &c); err != tc.err {
			t.Errorf("ParseFromMetadata() = %v; want %v", err, tc.err)
		}
	}
}

func TestDecodeOnly(t *testing.T) {
	secret := []byte("very-secret-secret")
	fn := DecodeOnly(secret)
	claims := Claims{
		Scope: "foo bar foobar",
	}
	token, err := Encode(claims, secret)
	if err != nil {
		t.Fatalf("Encode(%+v, %q) failed with: %v", claims, secret, err)
	}

	// good token
	md := metadata.New(map[string]string{"authorization": token})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	newCtx, err := fn(ctx)
	if err != nil {
		t.Errorf("Decode only should not be failed, but finally failed with: %v", err)
	}
	newClaims, ok := FromContext(newCtx)
	if !ok {
		t.Error("Claims missing from context")
	}
	if !reflect.DeepEqual(claims, newClaims) {
		t.Errorf("Claims not equal: %v vs %v", claims, newClaims)
	}

	// bad token
	md = metadata.New(map[string]string{"authorization": "bad token"})
	ctx = metadata.NewIncomingContext(context.Background(), md)
	newCtx, err = fn(ctx)
	if err != nil {
		t.Errorf("Decode only should not be failed, but finally failed with: %v", err)
	}
	newClaims, ok = FromContext(newCtx)
	if ok {
		t.Errorf("got claims=%v, want no claims", newClaims)
	}

	// no header
	newCtx, err = fn(context.Background())
	if err != nil {
		t.Errorf("Decode only should not be failed, but finally failed with: %v", err)
	}
	newClaims, ok = FromContext(newCtx)
	if ok {
		t.Errorf("got claims=%v, want no claims", newClaims)
	}
}

func TestTokenString(t *testing.T) {
	give := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJ0ZXN0In0.V9sFBcyxQYUeKTg_rzzjFqtL6b-x4TlDrmlgTTjMtkA"
	md := metadata.New(map[string]string{"authorization": give})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	// has header
	got := TokenString(ctx)
	if got != give {
		t.Errorf("got token=%s, want token=%s", got, give)
	}
	// no header
	give = ""
	got = TokenString(context.Background())
	if got != give {
		t.Errorf("got token=%s, want token=%s", got, give)
	}
}

func TestSubjectEqual(t *testing.T) {
	ctx := NewContext(context.Background(), Claims{
		Subject: "subject-1",
	})
	// equal
	if !SubjectEquals(ctx, "subject-1") {
		t.Errorf("got equal=false, want equal=true")
	}
	// not equal
	if SubjectEquals(ctx, "subject-2") {
		t.Errorf("got equal=true, want equal=false")
	}
}
