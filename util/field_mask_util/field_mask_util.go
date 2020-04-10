package field_mask_util

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

// ToSnakeCase transform the given string to snake_case.
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// TrimPrefix remove the prefix in the field mask.
func TrimPrefix(paths []string, prefix string) []string {
	rs := make([]string, 0)
	for _, p := range paths {
		rs = append(rs, strings.TrimPrefix(p, prefix))
	}
	return rs
}

// GetValidFieldMask return fields match with the value and definition of the given struct.
func GetValidFieldMask(paths []string, req interface{}) []string {
	npaths := make([]string, 0)
	for _, pth := range paths {
		if pth, ok := IsValidFieldMask(pth, req); ok {
			npaths = append(npaths, pth)
		}
	}
	return npaths
}

// IsValidFieldMask check whether the given path matchs with the defined paths
// in the given struct. It returns the normalized path follows snake_case format.
func IsValidFieldMask(path string, req interface{}) (string, bool) {
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
		nxtPath, ok := IsValidFieldMask(paths[1], v.Field(i).Interface())
		if !ok {
			return "", false
		}
		return fmt.Sprintf("%s.%s", npath, nxtPath), true
	}
	return npath, ok
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
