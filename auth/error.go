package auth

import "errors"

var (
	// ErrMetadataMissing reports that metadata is missing in the incoming context.
	ErrMetadataMissing = errors.New("auth: could not locate request metadata")
	// ErrAuthorizationMissing reports that authorization metadata is missing in the incoming context.
	ErrAuthorizationMissing = errors.New("auth: could not locate authorization metadata")
	//ErrInvalidToken reports that the token is invalid.
	ErrInvalidToken = errors.New("auth: invalid token")
	// ErrMultipleAuthFound reports that too many authorization entries were found.
	ErrMultipleAuthFound = errors.New("auth: too many authorization entries")
)
