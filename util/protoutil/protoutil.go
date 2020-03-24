package protoutil

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

// GetValidFieldMask return fields match with the value and definition of the given struct.
func GetValidFieldMask(pths []string, req interface{}) []string {
	npths := make([]string, 0)
	for _, pth := range pths {
		if ok, pth := IsValidFieldMask(pth, req); ok {
			npths = append(npths, pth)
		}
	}
	return npths
}

// IsValidFieldMask check whether the given path (pth) matchs with defined paths
// in the given struct. It returns the normalized path follows snake_case format.
func IsValidFieldMask(pth string, req interface{}) (bool, string) {
	fs := strings.SplitN(pth, ".", 2)
	var t reflect.Type
	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Struct {
		t = reflect.TypeOf(v.Interface())
	} else if v.Kind() == reflect.Ptr && !v.IsNil() {
		t = reflect.TypeOf(v.Elem().Interface())
	} else {
		return false, ""
	}
	ok := false
	cv := ""
	npth := ToSnakeCase(fs[0])
	i := 0
	for ; i < t.NumField(); i++ {
		// 1st priority default to be protobuf
		tag := t.Field(i).Tag.Get("protobuf")
		tagFields := strings.Split(tag, ",")
		cv = fmt.Sprintf("name=%s", ToSnakeCase(fs[0]))
		// 2nd priority json
		if tag == "" {
			tag = t.Field(i).Tag.Get("json")
			tagFields = strings.Split(tag, ",")
			cv = ToSnakeCase(fs[0])
		}
		//3rd priority field name
		if tag == "" {
			tag = ToSnakeCase(t.Field(i).Name)
			tagFields = []string{tag}
			cv = ToSnakeCase(fs[0])
		}
		for _, tagField := range tagFields {
			if tagField == cv {
				ok = true
				break
			}
		}
		if ok {
			break
		}
	}
	if !ok {
		return false, ""
	}
	if len(fs) > 1 { // has more fields?
		t = t.Field(i).Type
		var v interface{}
		fv := reflect.ValueOf(req).Field(i)
		if t.Kind() == reflect.Struct {
			v = fv.Interface()
		} else if t.Kind() == reflect.Ptr && !fv.IsNil() {
			v = fv.Elem().Interface()
		} else {
			return false, ""
		}
		valid, pth := IsValidFieldMask(fs[1], v)
		if valid {
			return valid, fmt.Sprintf("%s.%s", npth, pth)
		}
		return false, ""
	}
	return ok, npth
}
