package cache

import "errors"

var (
	ErrNotFound = errors.New("cache: key not found")
)
