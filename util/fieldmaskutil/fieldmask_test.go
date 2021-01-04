package fieldmaskutil_test

import (
	"testing"

	"github.com/pthethanh/micro/util/fieldmaskutil"
)

type (
	Meta struct {
		Values map[string]string
	}
	Location struct {
		X int `json:"x"`
		Y int `protobuf:"varint,2,opt,name=y,proto3" json:"y,omitempty"`
	}
	Address struct {
		Address  string
		Location Location
		Meta     []byte `protobuf:"bytes,2,opt,name=meta,proto3" json:"meta,omitempty"`
	}
	MyStruct struct {
		Name          string `json:"name"`
		Address       Address
		Meta          *Meta
		NextMeta      *Meta
		Metas         []*Meta
		SomeOtherMeta Meta `json:"some_other_meta"`
	}
)

func TestValidateFieldMask(t *testing.T) {
	v := MyStruct{
		NextMeta: &Meta{},
	}
	cases := []struct {
		name  string
		path  string
		valid bool
		npath string
	}{
		{
			"0", "wrong", false, "",
		},
		{
			"1", "name", true, "name",
		},
		{
			"2", "address.address", true, "address.address",
		},
		{
			"3", "address.location", true, "address.location",
		},
		{
			"4", "address.location.x", true, "address.location.x",
		},
		{
			"5", "address.location.Y", true, "address.location.y",
		},
		{
			"6", "address.Meta", true, "address.meta",
		},
		{
			"7", "meta", true, "meta",
		},
		{
			"nil pointer", "meta.values", false, "meta.values",
		},
		{
			"valid pointer", "NextMeta.values", true, "next_meta.values",
		},
		{
			"nil slice", "Metas", true, "metas",
		},
		{
			"wrong 1", "address.location.z", false, "",
		},
		{
			"snake case", "some_other_meta", true, "some_other_meta",
		},
		{
			"snake case 2", "SomeOtherMeta", true, "some_other_meta",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if npath, valid := fieldmaskutil.IsValid(c.path, v); valid != c.valid && npath != c.npath {
				t.Errorf("got valid=%t, path=%s, want valid=%t, path=%s", valid, npath, c.valid, c.path)
			}
		})
	}
}

func TestGetValidFieldMask(t *testing.T) {
	v := MyStruct{
		NextMeta: &Meta{},
	}
	paths := fieldmaskutil.GetValidFields([]string{
		"name",
		"NextMeta.values",
		"meta.values",
		"address.location.Y",
	}, v)
	want := 3
	if len(paths) != want {
		t.Errorf("got len(paths)=%d, want len(paths)=%d", len(paths), want)
	}
}

func TestTrimPrefix(t *testing.T) {
	v := MyStruct{
		Meta:     &Meta{},
		NextMeta: &Meta{},
	}
	paths := fieldmaskutil.GetValidFields([]string{
		"meta.values",
		"NextMeta.values",
	}, v, fieldmaskutil.TrimPrefix("meta."), fieldmaskutil.ToSnakeCase)
	want := 2
	if len(paths) != want {
		t.Fatalf("got len(paths)=%d, want len(paths)=%d", len(paths), want)
	}
	if paths[0] != "values" {
		t.Errorf("got field=%s, want field without prefix=%s", paths[0], "values")
	}
	if paths[1] != "next_meta.values" {
		t.Errorf("got field=%s, want field with snake_case=%s", paths[1], "next_meta.values")
	}
}

func TestRemoveFields(t *testing.T) {
	v := struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		Age      int    `json:"age"`
	}{
		Name:     "Jack",
		Password: "123",
		Age:      22,
	}
	paths := fieldmaskutil.GetValidFields([]string{
		"name",
		"password",
	}, v, fieldmaskutil.RemoveFields("password", "age"))
	want := 1
	if len(paths) != want {
		t.Errorf("got len(paths)=%d, want len(paths)=%d", len(paths), want)
	}
}

func TestContainsOneOf(t *testing.T) {
	cases := []struct {
		name   string
		in     []string
		v      []string
		expect bool
	}{
		{
			name:   "contain",
			in:     []string{"a", "b", "c"},
			v:      []string{"b"},
			expect: true,
		},
		{
			name:   "not contain",
			in:     []string{"a", "b", "c"},
			v:      []string{"d"},
			expect: false,
		},
		{
			name:   "contains",
			in:     []string{"a", "b", "c"},
			v:      []string{"e", "b"},
			expect: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if ok := fieldmaskutil.ContainsOneOf(c.in, c.v...); ok != c.expect {
				t.Errorf("got result=%v, want result=%v", ok, c.expect)
			}
		})
	}
}
