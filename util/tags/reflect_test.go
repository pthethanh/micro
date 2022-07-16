package tags_test

import (
	"errors"
	"sort"
	"testing"

	"github.com/pthethanh/micro/util/tags"
)

func TestTagsToStructFields(t *testing.T) {
	type Note struct {
		private string
		Value   string `json:"value,omitempty"`
	}
	type Address struct {
		Work string `json:"work,omitempty"`
		Home string `json:"home,omitempty"`
		Note Note   `json:"note,omitempty"`
	}
	v := struct {
		Name     string   `json:"name,omitempty"`
		Age      int      `json:"age,omitempty"`
		Address1 Address  `json:"address1,omitempty"`
		Address2 *Address `json:"address2,omitempty"`
		Chan     chan int
	}{
		Address2: &Address{},
	}
	// suppress unused fields
	_ = v.Address1.Note.private

	cases := []struct {
		name  string
		value interface{}
		tags  []string
		want  []string
	}{
		{
			name:  "some fields different levels",
			value: v,
			tags:  []string{"name", "age", "address1.work", "address2.note.value"},
			want:  []string{"Name", "Age", "Address1.Work", "Address2.Note.Value"},
		},
		{
			name:  "pointer, all fields - nil",
			value: &v,
			tags:  nil,
			want:  []string{"Name", "Age", "Address1.Note", "Address2", "Address2.Work", "Address2.Home", "Address2.Note.Value", "Address1", "Address1.Work", "Address1.Home", "Address1.Note.Value", "Address2.Note", "Chan"},
		},
		{
			name:  "all fields - *",
			value: v,
			tags:  []string{"*"},
			want:  []string{"Name", "Age", "Address1.Note", "Address2", "Address2.Work", "Address2.Home", "Address2.Note.Value", "Address1", "Address1.Work", "Address1.Home", "Address1.Note.Value", "Address2.Note", "Chan"},
		},
		{
			name:  "non-struct value",
			value: 2,
			tags:  nil,
			want:  []string{},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := tags.GetFieldNameMapping(c.value, "json", c.tags...)
			if err != nil {
				t.Error(err)
			}
			if len(got) != len(c.want) {
				t.Errorf("got fields=%s, want fields=%s", got, c.want)
			}
			values := got.Values()
			sort.Strings(values)
			sort.Strings(c.want)
			for i := 0; i < len(got); i++ {
				if values[i] != c.want[i] {
					t.Errorf("got fields=%s, want fields=%s", got, c.want)
					return
				}
			}
		})
	}
}

func TestTagsToTags(t *testing.T) {
	type Note struct {
		private string
		Value   string `json:"value,omitempty" bson:"bvalue,omitempty"`
	}
	type Address struct {
		Work string `json:"work,omitempty" bson:"bwork,omitempty"`
		Home string `json:"home,omitempty" bson:"bhome,omitempty"`
		Note Note   `json:"note,omitempty" bson:"bnote,omitempty"`
	}
	v := struct {
		Name     string   `json:"name,omitempty" bson:"bname,omitempty"`
		Age      int      `json:"age,omitempty" bson:"bage,omitempty"`
		Address1 Address  `json:"address1,omitempty" bson:"baddress1,omitempty"`
		Address2 *Address `json:"address2,omitempty" bson:"baddress2,omitempty"`
		Chan     int
	}{
		Address2: &Address{},
	}
	// suppress unused fields
	_ = v.Address1.Note.private
	cases := []struct {
		name  string
		value interface{}
		src   string
		dst   string
		tags  []string
		want  []string
	}{
		{
			name:  "some fields different levels with and without tag",
			value: v,
			src:   "json",
			dst:   "bson",
			tags:  []string{"name", "age", "address1.work", "address2.note.value", "Chan"},
			want:  []string{"bname", "bage", "baddress1.bwork", "baddress2.bnote.bvalue", "Chan"},
		},
		{
			name:  "pointer, all fields - nil",
			value: &v,
			tags:  nil,
			want:  []string{"bname", "bage", "baddress1.bnote", "baddress2", "baddress2.bwork", "baddress2.bhome", "baddress2.bnote.bvalue", "baddress1", "baddress1.bwork", "baddress1.bhome", "baddress1.bnote.bvalue", "baddress2.bnote", "Chan"},
		},
		{
			name:  "all fields - *",
			value: v,
			tags:  []string{"*"},
			want:  []string{"bname", "bage", "baddress1.bnote", "baddress2", "baddress2.bwork", "baddress2.bhome", "baddress2.bnote.bvalue", "baddress1", "baddress1.bwork", "baddress1.bhome", "baddress1.bnote.bvalue", "baddress2.bnote", "Chan"},
		},
		{
			name:  "non-struct value",
			value: 2,
			tags:  nil,
			want:  []string{},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := tags.GetTagMapping(c.value, "json", "bson", c.tags...)
			if err != nil {
				t.Error(err)
			}
			if len(got) != len(c.want) {
				t.Errorf("got tags=%s, want tags=%s", got, c.want)
			}
			values := got.Values()
			sort.Strings(values)
			sort.Strings(c.want)
			for i := 0; i < len(got); i++ {
				if values[i] != c.want[i] {
					t.Errorf("got tags=%s, want tags=%s", got, c.want)
					return
				}
			}
		})
	}
}

func TestTagsToFieldNamesNilValues(t *testing.T) {
	type Address struct {
		Home string `json:"home"`
		Work string `json:"work"`
	}
	v := struct {
		Name    *string  `json:"name"`
		Address *Address `json:"address"`
	}{
		// all nil
	}
	got, err := tags.GetFieldNameMapping(&v, "json", "name", "address", "address.home", "address.work")
	if err != nil {
		t.Error(err)
	}
	want := []string{"Name", "Address"}
	if len(got) != len(want) {
		t.Fatalf("got fields=%s, want fields=%s", got, want)
	}
	values := got.Values()
	sort.Strings(values)
	sort.Strings(want)
	for i := 0; i < len(got); i++ {
		if values[i] != want[i] {
			t.Errorf("got fields=%s, want fields=%s", got, want)
			return
		}
	}

	// pass nil value
	got, err = tags.GetFieldNameMapping(nil, "")
	if err != nil {
		t.Error(err)
	}
	if len(got) != 0 {
		t.Fatalf("pass nil value, got fields=%v, want fields=nil", got)
	}
}

func TestTagsToTagsNilValues(t *testing.T) {
	type Address struct {
		Home string `json:"home" bson:"bhome"`
		Work string `json:"work" bson:"bwork"`
	}
	v := struct {
		Name    *string  `json:"name" bson:"bname"`
		Address *Address `json:"address" bson:"baddress"`
	}{
		// all nil
	}
	got, err := tags.GetTagMapping(&v, "json", "bson", "name", "address", "address.home", "address.work")
	if err != nil {
		t.Error(err)
	}
	want := []string{"bname", "baddress"}
	if len(got) != len(want) {
		t.Fatalf("got tags=%s, want tags=%s", got, want)
	}
	values := got.Values()
	sort.Strings(values)
	sort.Strings(want)
	for i := 0; i < len(got); i++ {
		if values[i] != want[i] {
			t.Errorf("got tags=%s, want tags=%s", got, want)
			return
		}
	}

	// pass nil val
	got, err = tags.GetTagMapping(nil, "json", "bson")
	if err != nil {
		t.Error(err)
	}
	if len(got) != 0 {
		t.Fatalf("pass nil value, got tags=%v, want tags=nil", got)
	}
}

func TestResolverNotFound(t *testing.T) {
	if _, err := tags.GetFieldNameMapping(struct{}{}, "some_weird_tag", "test"); !errors.Is(err, tags.ErrResolverNotFound) {
		t.Errorf("got err=%v, want err=%v", err, tags.ErrResolverNotFound)
	}

	if _, err := tags.GetTagMapping(struct{}{}, "some_weird_src_tag", "json"); !errors.Is(err, tags.ErrResolverNotFound) {
		t.Errorf("got err=%v, want err=%v", err, tags.ErrResolverNotFound)
	}
	if _, err := tags.GetTagMapping(struct{}{}, "json", "some_weird_dst_tag"); !errors.Is(err, tags.ErrResolverNotFound) {
		t.Errorf("got err=%v, want err=%v", err, tags.ErrResolverNotFound)
	}
}
