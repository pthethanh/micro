// Package validator provides convenient utilities for validation using https://github.com/go-playground/validator.
package validator

import (
	"sync"

	validate "github.com/go-playground/validator/v10"
)

var (
	once      sync.Once
	validator *validate.Validate
)

// New return instance of validator
func New() *validate.Validate {
	once.Do(func() {
		validator = validate.New()
	})
	return validator
}

// Validate a structs exposed fields base on the definition of 'validate' tag.
func Validate(v interface{}) error {
	return New().Struct(v)
}

// ValidatePartial validates the fields passed in only, ignoring all others.
func ValidatePartial(v interface{}, fields ...string) error {
	return New().StructPartial(v, fields...)
}

// ValidateExcept validates all the fields except the given fields.
func ValidateExcept(v interface{}, fields ...string) error {
	return New().StructExcept(v, fields...)
}

// Var validates a single variable using tag style validation.
func Var(field interface{}, tag string) error {
	return New().Var(field, tag)
}

// RegisterValidation adds a validation with the given tag
func RegisterValidation(tag string, fn validate.Func, callValidationEvenIfNull bool) error {
	return New().RegisterValidation(tag, fn, callValidationEvenIfNull)
}
