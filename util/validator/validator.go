// Package validator provides convenient utilities for validation.
package validator

import (
	"sync"

	validate "github.com/go-playground/validator/v10"
)

var (
	once      sync.Once
	validator *validate.Validate
)

// New return new instance of validator with the given tag.
func New(tag string) *validate.Validate {
	v := validate.New()
	if tag != "" {
		v.SetTagName(tag)
	}
	return v
}

// Root return root validator instance using 'validate' tag.
func Root() *validate.Validate {
	once.Do(func() {
		validator = New("")
	})
	return validator
}

// Validate a struct exposed fields base on the definition of 'validate' tag.
func Validate(v interface{}) error {
	return Root().Struct(v)
}

// ValidatePartial validates the fields passed in only, ignoring all others.
func ValidatePartial(v interface{}, fields ...string) error {
	return Root().StructPartial(v, fields...)
}

// ValidateExcept validates all the fields except the given fields.
func ValidateExcept(v interface{}, fields ...string) error {
	return Root().StructExcept(v, fields...)
}

// Var validates a single variable using tag style validation.
func Var(field interface{}, tag string) error {
	return Root().Var(field, tag)
}

// RegisterValidation adds a validation with the given tag
func RegisterValidation(tag string, fn validate.Func, callValidationEvenIfNull bool) error {
	return Root().RegisterValidation(tag, fn, callValidationEvenIfNull)
}
