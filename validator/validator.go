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

// Validate the given struct base on the definition of 'validate' tag of the struct.
func Validate(v interface{}) error {
	return New().Struct(v)
}

// Var validates a single variable using tag style validation.
func Var(field interface{}, tag string) error {
	return New().Var(field, tag)
}

// RegisterValidation adds a validation with the given tag
//
// NOTES:
// - if the key already exists, the previous validation function will be replaced.
// - this method is not thread-safe it is intended that these all be registered prior to any validation
func RegisterValidation(tag string, fn validate.Func, callValidationEvenIfNull bool) error {
	return New().RegisterValidation(tag, fn, callValidationEvenIfNull)
}
