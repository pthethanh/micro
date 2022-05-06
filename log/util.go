package log

import (
	"encoding/json"
	"fmt"
)

// MustJSON return JSON string of the given value.
// In case of a marshal failure, return string value using fmt package.
func MustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
