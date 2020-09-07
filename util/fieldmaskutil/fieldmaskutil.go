// Package fieldmaskutil provides convenient utilities for working with field_mask.
package fieldmaskutil

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

type (
	// TransformFunc is a filed path transform func.
	TransformFunc = func(string) string
)

// ToSnakeCase transform the given string to snake_case.
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// TrimPrefix is a transform func to remove the prefix of the field.
func TrimPrefix(prefix string) TransformFunc {
	return func(s string) string {
		return strings.TrimPrefix(s, prefix)
	}
}

// RemoveFields is an option to remove a specific fields out of the field mask result.
// This can be used to remove sensitive fields out of the update.
func RemoveFields(fields ...string) TransformFunc {
	set := make(map[string]struct{})
	for _, f := range fields {
		set[f] = struct{}{}
	}
	return func(s string) string {
		// if the field is specified, return empty so that it will be ignored in the result.
		if _, ok := set[s]; ok {
			return ""
		}
		return s
	}
}

// GetValidFields return fields match with the value and definition of the given struct.
// Transformation rules on the result path can be given via transform options.
func GetValidFields(paths []string, req interface{}, opts ...TransformFunc) []string {
	npaths := make([]string, 0)
	for _, pth := range paths {
		if pth, ok := IsValid(pth, req); ok {
			for _, opt := range opts {
				pth = opt(pth)
			}
			if pth != "" {
				npaths = append(npaths, pth)
			}
		}
	}
	return npaths
}

// IsValid check whether the given path matchs with the defined paths
// in the given struct. It returns the normalized path follows snake_case format.
func IsValid(path string, req interface{}) (string, bool) {
	if req == nil {
		return "", false
	}
	paths := strings.SplitN(path, ".", 2)
	t := reflect.TypeOf(req)
	v := reflect.ValueOf(req)
	if t.Kind() == reflect.Ptr && v.IsValid() && !v.IsNil() {
		t = reflect.TypeOf(v.Elem().Interface())
		v = reflect.ValueOf(v.Elem().Interface())
	} else if t.Kind() != reflect.Struct {
		return "", false
	}
	ok := false
	npath := ToSnakeCase(paths[0])
	i := 0
	for ; i < t.NumField(); i++ {
		if isValidTag(t.Field(i), npath) {
			ok = true
			break
		}
	}
	if !ok {
		return "", false
	}
	if len(paths) > 1 {
		nxtPath, ok := IsValid(paths[1], v.Field(i).Interface())
		if !ok {
			return "", false
		}
		return fmt.Sprintf("%s.%s", npath, nxtPath), true
	}
	return npath, ok
}

// ContainsOneOf check if the list contains one of the provided fields.
func ContainsOneOf(list []string, fields ...string) bool {
	s := make(map[string]struct{})
	for _, l := range list {
		s[l] = struct{}{}
	}
	for _, f := range fields {
		if _, ok := s[f]; ok {
			return true
		}
	}
	return false
}

// check if the tag name is mentioned in the tags of the given field.
// the priority to check is protobuf, json and field name.
func isValidTag(f reflect.StructField, tagName string) bool {
	tag := f.Tag.Get("protobuf")
	tagToks := strings.Split(tag, ",")
	snakeTagName := ToSnakeCase(tagName)
	v := fmt.Sprintf("name=%s", snakeTagName)
	if tag == "" {
		tag = f.Tag.Get("json")
		tagToks = strings.Split(tag, ",")
		v = snakeTagName
	}
	if tag == "" {
		tag = ToSnakeCase(f.Name)
		tagToks = []string{tag}
		v = snakeTagName
	}
	for _, tok := range tagToks {
		if tok == v {
			return true
		}
	}
	return false
}
