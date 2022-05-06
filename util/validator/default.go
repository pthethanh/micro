package validator

import "sync"

var (
	cache sync.Map
)

// Get get the tag validator from cache or create new one if not exists.
// If tag is empty, return Root() validator.
func Get(tag string) *Validator {
	if tag == "" {
		return Root()
	}
	if v, ok := cache.Load(tag); ok {
		return v.(*Validator)
	}
	v := New(tag)
	cache.Store(tag, v)
	return v
}

// Validate a struct exposed fields base on the definition of validate tag.
func Validate(v any) error {
	return Root().Validate(v)
}

// ValidatePartial validates the fields passed in only, ignoring all others.
func ValidatePartial(v any, fields ...string) error {
	return Root().ValidatePartial(v, fields...)
}

// ValidateExcept validates all the fields except the given fields.
func ValidateExcept(v any, fields ...string) error {
	return Root().ValidateExcept(v, fields...)
}

// Var validates a single variable using tag style validation.
func Var(field any, tag string) error {
	return Root().Var(field, tag)
}
