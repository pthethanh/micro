package httputil_test

import (
	"net/url"
	"testing"

	"github.com/pthethanh/micro/util/httputil"
)

func TestDecodeQuery(t *testing.T) {
	type employee struct {
		Name string
		Age  uint
	}
	v := employee{}
	values := url.Values{
		"name": []string{"jack"},
		"age":  []string{"22"},
	}
	err := httputil.DecodeQuery(&v, values)
	check := func(v *employee, err error) {
		if err != nil {
			t.Fatal(err)
		}
		if v.Name != "jack" {
			t.Errorf("got name=%s, want name=jack", v.Name)
		}
		if v.Age != 22 {
			t.Errorf("got age=%d, want age=22", v.Age)
		}
	}
	check(&v, err)

	// using decoder
	err = httputil.NewQueryDecoder("query").Decode(&v, values)
	check(&v, err)
}

func TestEncodeQuery(t *testing.T) {
	type employee struct {
		Name string `query:"name"`
		Age  uint   `query:"age"`
	}
	give := employee{
		Name: "jack",
		Age:  22,
	}
	got := url.Values{}
	err := httputil.EncodeQuery(give, got)
	check := func(got url.Values, err error) {
		if err != nil {
			t.Fatal(err)
		}
		want := "age=22&name=jack"
		if got.Encode() != want {
			t.Errorf("got query=%s, want query=%s", got.Encode(), want)
		}
	}
	check(got, err)

	// using encoder
	got = url.Values{}
	err = httputil.NewQueryEncoder("query").Encode(give, got)
	check(got, err)
}
