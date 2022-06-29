package reflectutil

import (
	"encoding/json"
)

// ToMap convert a struct to a map.
func ToMap[K comparable, V interface{}](v interface{}) (map[K]V, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	m := make(map[K]V)
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
