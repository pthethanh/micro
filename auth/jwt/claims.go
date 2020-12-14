package jwt

import (
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// Claims represents the claims provided by the JWT.
type Claims struct {
	// Auth claims
	Audience  string `json:"aud,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	ID        string `json:"jti,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	NotBefore int64  `json:"nbf,omitempty"`
	Subject   string `json:"sub,omitempty"`

	// User attributes claims
	Name                string `json:"name,omitempty"`
	GivenName           string `json:"given_name,omitempty"`
	FamilyName          string `json:"family_name,omitempty"`
	MiddleName          string `json:"middle_name,omitempty"`
	Nickname            string `json:"nickname,omitempty"`
	PreferredUsername   string `json:"preferred_username,omitempty"`
	Profile             string `json:"profile,omitempty"`
	Picture             string `json:"picture,omitempty"`
	Website             string `json:"website,omitempty"`
	Email               string `json:"email,omitempty"`
	EmailVerified       bool   `json:"email_verified,omitempty"`
	Gender              string `json:"gender,omitempty"`
	Birthdate           string `json:"birthdate,omitempty"`
	Zoneinfo            string `json:"zoneinfo,omitempty"`
	Locale              string `json:"locale,omitempty"`
	PhoneNumber         string `json:"phone_number,omitempty"`
	PhoneNumberVerified bool   `json:"phone_number_verified,omitempty"`
	Address             string `json:"address,omitempty"`
	UpdatedAt           int64  `json:"updated_at,omitempty"`

	// Custom attributes claims.
	Scope    string                 `json:"scope,omitempty"`
	Admin    bool                   `json:"admin,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ContainScopes checks if `scopes` are present within the Claim.Scope.
func (c Claims) ContainScopes(scopes ...string) bool {
	currentScopes := strings.Split(c.Scope, " ")
	if len(currentScopes) == 0 {
		return false
	}
	for _, scope := range scopes {
		match := false
		for _, s := range currentScopes {
			if scope == s {
				match = true
			}
		}
		if !match {
			return false
		}
	}
	return true
}

// Valid implement jwt.Claims interface.
func (c Claims) Valid() error {
	claims := jwt.StandardClaims{
		Id:        c.ID,
		Audience:  c.Audience,
		ExpiresAt: c.ExpiresAt,
		IssuedAt:  c.IssuedAt,
		Issuer:    c.Issuer,
		NotBefore: c.NotBefore,
		Subject:   c.Subject,
	}
	return claims.Valid()
}
