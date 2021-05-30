// Package reflectutil provides some convenient utilities for working with reflect.
package reflectutil

import (
	"reflect"
)

// GetFieldNamesFromTags return struct' field names of the given tag's values.
// Return all field names if values is nil or its first value is *.
// Return nil if the given value is not a struct.
func GetFieldNamesFromTags(v interface{}, tag string, values ...string) []string {
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

// GetTagsFromTags get tag mapping values coresponding to the given tag values.
func GetTagsFromTags(v interface{}, srcTag, dstTag string, values ...string) []string {
	m := make(map[string]string)
	tagsToTags(m, v, "", "", srcTag, dstTag)
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

// tagsToFieldNames write the mapping between tags and field names to the given res map.
func tagsToTags(res map[string]string, req interface{}, tag1Prefix string, tag2Prefix, tag1, tag2 string) {
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
		res[tag1Prefix+t.Field(i).Tag.Get(tag1)] = tag2Prefix + t.Field(i).Tag.Get(tag2)
		if fv.Kind() == reflect.Struct {
			tag1Prefix := tag1Prefix + t.Field(i).Tag.Get(tag1) + "."
			tag2Prefix := tag2Prefix + t.Field(i).Tag.Get(tag2) + "."
			tagsToTags(res, fv.Interface(), tag1Prefix, tag2Prefix, tag1, tag2)
		}
	}
}
