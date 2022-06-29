package log_test

import (
	"testing"

	"github.com/pthethanh/micro/log"
)

func TestMustJSON(t *testing.T) {
	cases := []struct {
		name string
		v    interface{}
		want string
	}{
		{
			name: "string",
			v:    "test",
			want: `"test"`,
		},
		{
			name: "number",
			v:    1,
			want: `1`,
		},
		{
			name: "map",
			v:    map[string]int{"hcm": 1},
			want: `{"hcm":1}`,
		},
		{
			name: "struct",
			v: struct {
				Name string
			}{
				Name: "jack",
			},
			want: `{"Name":"jack"}`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := log.MustJSON(c.v); got != c.want {
				t.Errorf("got result=%v, want result=%v\n", got, c.want)
			}
		})
	}
}
