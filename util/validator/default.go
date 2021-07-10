package validator

import "sync"

var (
	cache sync.Map
)

// Get get the tag validator from cache or create new one if not exists.
// If tag is empty, return Root() validator.
func Get(tag string) *Validator {
	if tag == "" {
		return Root()
	}
	if v, ok := cache.Load(tag); ok {
		return v.(*Validator)
	}
	v := New(tag)
	cache.Store(tag, v)
	return v
}
