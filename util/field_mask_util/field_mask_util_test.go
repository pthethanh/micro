package field_mask_util_test

import (
	"testing"

	"gitlab.com/akoala/micro/util/field_mask_util"
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
			if npath, valid := field_mask_util.IsValidFieldMask(c.path, v); valid != c.valid && npath != c.npath {
				t.Errorf("got valid=%t, path=%s, want valid=%t, path=%s", valid, npath, c.valid, c.path)
			}
		})
	}
}

func TestGetValidFieldMask(t *testing.T) {
	v := MyStruct{
		NextMeta: &Meta{},
	}
	paths := field_mask_util.GetValidFieldMask([]string{
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
	rs := field_mask_util.TrimPrefix([]string{"comment.status", "comment.content"}, "comment.")
	if rs[0] != "status" || rs[1] != "content" {
		t.Fatalf("got rs=%v, want rs=%v", rs, []string{"status", "content"})
	}
}
