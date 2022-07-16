// Package tags provides some convenient utilities for working with reflect.
// Note that this package is just an experiment and might be removed in the future.
package tags

import (
	"errors"
	"reflect"
	"strings"
	"unicode"

	"github.com/pthethanh/micro/util/maps"
)

type (
	// ResolverFunc is a function to resolve tag value.
	ResolverFunc = func(string) string

	// GetFieldNamesFromTagsRequest hold request information
	// for getting field names from tags' values.
	GetFieldNamesFromTagsRequest struct {
		Value     interface{}
		Tag       string
		Resolver  ResolverFunc
		TagValues []string
	}

	// GetTagMappingRequest hold request information
	// for getting tags' values from another tags' values.
	GetTagMappingRequest struct {
		Value       interface{}
		SrcTag      string
		SrcResolver ResolverFunc
		DstTag      string
		DstResolver ResolverFunc
		TagValues   []string
	}
)

var (
	ErrResolverNotFound = errors.New("resolver not found, register new resolver with tags.RegisterResolver")

	resolvers = map[string]ResolverFunc{}
)

func init() {
	firstValFunc := func(v string) string {
		return strings.Split(v, ",")[0]
	}
	protoFunc := func(v string) string {
		parts := strings.Split(v, ",")
		for _, part := range parts {
			if strings.HasPrefix(part, "name=") {
				return strings.TrimPrefix(part, "name=")
			}
		}
		return v
	}
	RegisterResolver("json", firstValFunc)
	RegisterResolver("xml", firstValFunc)
	RegisterResolver("bson", firstValFunc)
	RegisterResolver("db", firstValFunc)
	RegisterResolver("protobuf", protoFunc)
}

// RegisterResolver register a new resolver func for a specific tag.
func RegisterResolver(name string, r ResolverFunc) {
	resolvers[name] = r
}

// GetFieldNameMapping return struct's field names mapping of the given tags' values.
// Return all field names if tags' values are nil or its first value is *.
// Return empty map if the given structVal is not a struct.
func GetFieldNameMapping(structVal interface{}, tag string, tagValues ...string) (maps.Map[string, string], error) {
	if structVal == nil {
		return make(maps.Map[string, string]), nil
	}
	m := make(maps.Map[string, string])
	r, ok := resolvers[tag]
	if !ok {
		return nil, ErrResolverNotFound
	}

	tagsToFieldNames(m, structVal, "", "", tag, r)
	addAll := len(tagValues) == 0 || tagValues[0] == "*"
	rs := make(maps.Map[string, string])
	if addAll {
		for k, v := range m {
			rs[k] = v
		}
		return rs, nil
	}
	for _, v := range tagValues {
		if fv, ok := m[v]; ok {
			rs[v] = fv
		}
	}
	return rs, nil
}

// GetTagMapping get tag mapping values corresponding to the given tag values.
func GetTagMapping(structVal interface{}, srcTag string, dstTag string, srcTagValues ...string) (maps.Map[string, string], error) {
	if structVal == nil {
		return make(maps.Map[string, string]), nil
	}
	m := make(maps.Map[string, string])
	SrcResolver, ok := resolvers[srcTag]
	if !ok {
		return nil, ErrResolverNotFound
	}
	DstResolver, ok := resolvers[dstTag]
	if !ok {
		return nil, ErrResolverNotFound
	}
	tagsToTags(m, structVal, "", srcTag, SrcResolver, "", dstTag, DstResolver)
	addAll := len(srcTagValues) == 0 || srcTagValues[0] == "*"
	rs := make(maps.Map[string, string])
	if addAll {
		for k, v := range m {
			rs[k] = v
		}
		return rs, nil
	}
	for _, v := range srcTagValues {
		if fv, ok := m[v]; ok {
			rs[v] = fv
		}
	}
	return rs, nil
}

// tagsToFieldNames write the mapping between tags and field names to the given res map.
func tagsToFieldNames(res maps.Map[string, string], req interface{}, namePrefix string, tagPrefix, tag string, resolver ResolverFunc) {
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
func tagsToTags(res maps.Map[string, string], req interface{}, prefix1 string, tag1 string, resolver1 ResolverFunc, prefix2, tag2 string, resolver2 ResolverFunc) {
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
