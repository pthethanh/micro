// Package validator provides convenient utilities for validation using https://github.com/go-playground/validator.
package validator

import (
	"reflect"
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

// ValidatePartial validates all the fields except the given fields.
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

// TagsToFieldNames return struct's field names of the given tag's values.
// Return all field names if values is nil or its first value is *.
// Return nil if the given value is not a struct.
func TagsToFieldNames(v interface{}, tag string, values ...string) []string {
	m := make(map[string]string)
	tagsToFieldNames(m, v, "", "", tag)
	addAll := len(values) == 0 || values[0] == "*"
	rs := make([]string, 0)
	if addAll {
		for _, v := range m {
			rs = append(rs, v)
		}
		return rs
	}
	for _, v := range values {
		if fv, ok := m[v]; ok {
			rs = append(rs, fv)
		}
	}
	return rs
}

// tagsToFieldNames write the mapping between tags and field names to the given res map.
func tagsToFieldNames(res map[string]string, req interface{}, namePrefix string, tagPrefix, tag string) {
	t := reflect.TypeOf(req)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < t.NumField(); i++ {
		fv := reflect.ValueOf(req).Field(i)
		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}
		res[tagPrefix+t.Field(i).Tag.Get(tag)] = namePrefix + t.Field(i).Name
		if fv.Kind() == reflect.Struct {
			fPrefix := namePrefix + t.Field(i).Name + "."
			tPrefix := tagPrefix + t.Field(i).Tag.Get(tag) + "."
			tagsToFieldNames(res, fv.Interface(), fPrefix, tPrefix, tag)
		}
	}
}
