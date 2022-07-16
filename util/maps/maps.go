// Package maps provides some convenient utilities for working with maps.
// Note that this package is just an experiment and might be removed in the future.
package maps

import (
	"encoding/json"
	"reflect"
)

type (
	// Map is a convenient map utilities.
	Map[K comparable, V any] map[K]V
)

// New convert a struct to a map.
func New[K comparable, V any](v any) (Map[K, V], error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	m := make(Map[K, V])
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// Equal return true if 2 maps equal.
func (m Map[K, V]) Equal(m1 Map[K, V]) bool {
	if len(m) != len(m1) {
		return false
	}
	for k, v1 := range m {
		if v2, ok := m1[k]; !ok || !reflect.DeepEqual(v1, v2) {
			return false
		}
	}
	return true
}

// Clear removes all entries from m, leaving it empty.
// Return the original map.
func (m Map[K, V]) Clear() Map[K, V] {
	for k := range m {
		delete(m, k)
	}
	return m
}

// Clone returns a copy of m.  This is a shallow clone:
// the new keys and values are set using ordinary assignment.
func (m Map[K, V]) Clone() Map[K, V] {
	r := make(Map[K, V], len(m))
	for k, v := range m {
		r[k] = v
	}
	return r
}

// Copy copies all key/value pairs in src adding them to dst.
// When a key in src is already present in dst,
// the value in dst will be overwritten by the value associated
// with the key in src.
func (m Map[K, V]) Copy(dst Map[K, V]) {
	for k, v := range m {
		dst[k] = v
	}
}

// Keys return all keys of the map.
func (m Map[K, V]) Keys() []K {
	rs := make([]K, len(m))
	i := 0
	for k := range m {
		rs[i] = k
		i++
	}
	return rs
}

// Values return values of the map.
func (m Map[K, V]) Values() []V {
	rs := make([]V, len(m))
	i := 0
	for _, v := range m {
		rs[i] = v
		i++
	}
	return rs
}

// Set set the given value to the given keys.
// If the given keys empty, set the given value to all keys.
// Return the original map.
func (m Map[K, V]) Set(v V, ks ...K) Map[K, V] {
	if len(ks) == 0 {
		for k := range m {
			m[k] = v
		}
		return m
	}
	for _, k := range ks {
		m[k] = v
	}
	return m
}

// Filter filter and transform the map to a new map base on the given filter & transform function.
func (m Map[K, V]) Filter(f func(k K, v V) (bool, K, V)) Map[K, V] {
	rs := make(Map[K, V])
	for k, v := range m {
		ok, k1, v1 := f(k, v)
		if ok {
			rs[k1] = v1
		}
	}
	return rs
}

// Delete delete the given keys and return the original map.
func (m Map[K, V]) Delete(ks ...K) Map[K, V] {
	for _, k := range ks {
		delete(m, k)
	}
	return m
}
