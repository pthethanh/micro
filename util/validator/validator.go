// Package validator provides convenient utilities for validation.
package validator

import (
	validate "github.com/go-playground/validator/v10"
)

type (
	// Validator is a validation helper.
	Validator struct {
		v *validate.Validate
	}
)

var (
	root *Validator
)

// New return new instance of validator with the given tag.
func New(tag string) *Validator {
	v := validate.New()
	if tag != "" {
		v.SetTagName(tag)
	}
	return &Validator{
		v: v,
	}
}

// Init inits the root validator with the given tag.
func Init(tag string) *Validator {
	root = New(tag)
	return root
}

// Root return root validator instance using default 'validate' tag.
func Root() *Validator {
	if root == nil {
		root = Init("")
	}
	return root
}

// Validate a struct exposed fields base on the definition of validate tag.
func (validator *Validator) Validate(v any) error {
	return validator.v.Struct(v)
}

// ValidatePartial validates the fields passed in only, ignoring all others.
func (validator *Validator) ValidatePartial(v any, fields ...string) error {
	return validator.v.StructPartial(v, fields...)
}

// ValidateExcept validates all the fields except the given fields.
func (validator *Validator) ValidateExcept(v any, fields ...string) error {
	return validator.v.StructExcept(v, fields...)
}

// Var validates a single variable using tag style validation.
func (validator *Validator) Var(field any, tag string) error {
	return validator.v.Var(field, tag)
}

// Register adds a validation with the given tag
func (validator *Validator) Register(tag string, fn validate.Func, callValidationEvenIfNull bool) error {
	return validator.v.RegisterValidation(tag, fn, callValidationEvenIfNull)
}
