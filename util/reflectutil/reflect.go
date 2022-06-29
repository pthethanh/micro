// Package reflectutil provides some convenient utilities for working with reflect.
// Note that this package is just an experiment and might be removed in the future.
package reflectutil

import (
	"reflect"
	"strings"
	"unicode"
)

type (
	// TagResolverFunc is a function to resolve tag value.
	TagResolverFunc = func(string) string

	// GetFieldNamesFromTagsRequest hold request information
	// for getting field names from tags' values.
	GetFieldNamesFromTagsRequest struct {
		Value     interface{}
		Tag       string
		Resolver  TagResolverFunc
		TagValues []string
	}

	// GetTagsFromTagsRequest hold request information
	// for getting tags' values from another tags' values.
	GetTagsFromTagsRequest struct {
		Value       interface{}
		SrcTag      string
		SrcResolver TagResolverFunc
		DstTag      string
		DstResolver TagResolverFunc
		TagValues   []string
	}
)

var (
	// FirstValueTagResolverFunc is a tag resolver return first value after split by comma.
	FirstValueTagResolverFunc = func(v string) string {
		return strings.Split(v, ",")[0]
	}

	// ProtobufTagResolverFunc is a tag resolver for resolving protobuf tag name.
	ProtobufTagResolverFunc = func(v string) string {
		parts := strings.Split(v, ",")
		for _, part := range parts {
			if strings.HasPrefix(part, "name=") {
				return strings.TrimPrefix(part, "name=")
			}
		}
		return v
	}
)

// GetFieldNamesFromTags return struct' field names of the given tag's values.
// Return all field names if values is nil or its first value is *.
// Return nil if the given value is not a struct.
func GetFieldNamesFromTags(req GetFieldNamesFromTagsRequest) map[string]string {
	if req.Value == nil {
		return nil
	}
	m := make(map[string]string)
	if req.Resolver == nil {
		req.Resolver = FirstValueTagResolverFunc
	}
	tagsToFieldNames(m, req.Value, "", "", req.Tag, req.Resolver)
	addAll := len(req.TagValues) == 0 || req.TagValues[0] == "*"
	rs := make(map[string]string)
	if addAll {
		for k, v := range m {
			rs[k] = v
		}
		return rs
	}
	for _, v := range req.TagValues {
		if fv, ok := m[v]; ok {
			rs[v] = fv
		}
	}
	return rs
}

// GetTagsFromTags get tag mapping values corresponding to the given tag values.
func GetTagsFromTags(req GetTagsFromTagsRequest) map[string]string {
	if req.Value == nil {
		return nil
	}
	m := make(map[string]string)
	if req.SrcResolver == nil {
		req.SrcResolver = FirstValueTagResolverFunc
	}
	if req.DstResolver == nil {
		req.DstResolver = FirstValueTagResolverFunc
	}
	tagsToTags(m, req.Value, "", req.SrcTag, req.SrcResolver, "", req.DstTag, req.DstResolver)
	addAll := len(req.TagValues) == 0 || req.TagValues[0] == "*"
	rs := make(map[string]string)
	if addAll {
		for k, v := range m {
			rs[k] = v
		}
		return rs
	}
	for _, v := range req.TagValues {
		if fv, ok := m[v]; ok {
			rs[v] = fv
		}
	}
	return rs
}

// tagsToFieldNames write the mapping between tags and field names to the given res map.
func tagsToFieldNames(res map[string]string, req interface{}, namePrefix string, tagPrefix, tag string, resolver TagResolverFunc) {
	if req == nil {
		return
	}
	t := reflect.TypeOf(req)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}
	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		// ignore un-exported fields.
		if unicode.IsLower(rune(t.Field(i).Name[0])) {
			continue
		}
		fv := v.Field(i)
		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}
		tagValue := resolver(t.Field(i).Tag.Get(tag))
		if tagValue == "" {
			tagValue = t.Field(i).Name
		}
		tagValue = tagPrefix + tagValue
		nameValue := namePrefix + t.Field(i).Name
		res[tagValue] = nameValue
		if fv.Kind() == reflect.Struct {
			tagPrefix := tagValue + "."
			namePrefix := nameValue + "."
			tagsToFieldNames(res, fv.Interface(), namePrefix, tagPrefix, tag, resolver)
		}
	}
}

// tagsToFieldNames write the mapping between tags and field names to the given res map.
func tagsToTags(res map[string]string, req interface{}, prefix1 string, tag1 string, resolver1 TagResolverFunc, prefix2, tag2 string, resolver2 TagResolverFunc) {
	if req == nil {
		return
	}
	t := reflect.TypeOf(req)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}
	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		// ignore un-exported fields.
		if unicode.IsLower(rune(t.Field(i).Name[0])) {
			continue
		}
		fv := v.Field(i)
		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}
		v1 := resolver1(t.Field(i).Tag.Get(tag1))
		if v1 == "" {
			v1 = t.Field(i).Name
		}
		v2 := resolver2(t.Field(i).Tag.Get(tag2))
		if v2 == "" {
			v2 = t.Field(i).Name
		}
		v1 = prefix1 + v1
		v2 = prefix2 + v2
		res[v1] = v2
		if fv.Kind() == reflect.Struct {
			prefix1 := v1 + "."
			prefix2 := v2 + "."
			tagsToTags(res, fv.Interface(), prefix1, tag1, resolver1, prefix2, tag2, resolver2)
		}
	}
}
