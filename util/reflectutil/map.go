package reflectutil

import (
	"encoding/json"
	"fmt"
)

type (
	// JSONObject is a map prepresent a struct information.
	JSONObject map[string]interface{}
)

// ToJSONObject convert a struct to a json object/map.
func ToJSONObject(v interface{}) (JSONObject, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// Keys return keys of the json object.
func (m JSONObject) Keys() []string {
	v := make([]string, 0)
	for k := range m {
		v = append(v, k)
	}
	return v
}

// Values return values of the json object.
func (m JSONObject) Values() []interface{} {
	v := make([]interface{}, 0)
	for _, vv := range m {
		v = append(v, vv)
	}
	return v
}

// StringValues return values of the json object as strings.
func (m JSONObject) StringValues() []string {
	v := make([]string, 0)
	for _, vv := range m {
		v = append(v, fmt.Sprintf("%v", vv))
	}
	return v
}

// Sets set value of the keys to the given value.
func (m JSONObject) Sets(keys []string, value interface{}) {
	for _, k := range keys {
		m[k] = value
	}
}
