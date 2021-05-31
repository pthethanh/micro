package reflectutil_test

import (
	"sort"
	"testing"

	"github.com/pthethanh/micro/util/reflectutil"
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
			got := reflectutil.GetFieldNamesFromTags(reflectutil.GetFieldNamesFromTagsRequest{
				Value:     c.value,
				Tag:       "json",
				Resolver:  reflectutil.JSONTagResolverFunc,
				TagValues: c.tags,
			})
			if len(got) != len(c.want) {
				t.Errorf("got fields=%s, want fields=%s", got, c.want)
			}
			sort.Strings(got)
			sort.Strings(c.want)
			for i := 0; i < len(got); i++ {
				if got[i] != c.want[i] {
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
			got := reflectutil.GetTagsFromTags(reflectutil.GetTagsFromTagsRequest{
				Value:       c.value,
				SrcTag:      "json",
				SrcResolver: reflectutil.JSONTagResolverFunc,
				DstTag:      "bson",
				DstResolver: reflectutil.FirstValueTagResolverFunc,
				TagValues:   c.tags,
			})
			if len(got) != len(c.want) {
				t.Errorf("got tags=%s, want tags=%s", got, c.want)
			}
			sort.Strings(got)
			sort.Strings(c.want)
			for i := 0; i < len(got); i++ {
				if got[i] != c.want[i] {
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
	got := reflectutil.GetFieldNamesFromTags(reflectutil.GetFieldNamesFromTagsRequest{
		Value:     &v,
		Tag:       "json",
		Resolver:  reflectutil.JSONTagResolverFunc,
		TagValues: []string{"name", "address", "address.home", "address.work"},
	})
	want := []string{"Name", "Address"}
	if len(got) != len(want) {
		t.Fatalf("got fields=%s, want fields=%s", got, want)
	}
	sort.Strings(got)
	sort.Strings(want)
	for i := 0; i < len(got); i++ {
		if got[i] != want[i] {
			t.Errorf("got fields=%s, want fields=%s", got, want)
			return
		}
	}

	// pass nil value
	got = reflectutil.GetFieldNamesFromTags(reflectutil.GetFieldNamesFromTagsRequest{
		Value: nil,
	})
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
	got := reflectutil.GetTagsFromTags(reflectutil.GetTagsFromTagsRequest{
		Value:       &v,
		SrcTag:      "json",
		SrcResolver: reflectutil.JSONTagResolverFunc,
		DstTag:      "bson",
		DstResolver: reflectutil.FirstValueTagResolverFunc,
		TagValues:   []string{"name", "address", "address.home", "address.work"},
	})
	want := []string{"bname", "baddress"}
	if len(got) != len(want) {
		t.Fatalf("got tags=%s, want tags=%s", got, want)
	}
	sort.Strings(got)
	sort.Strings(want)
	for i := 0; i < len(got); i++ {
		if got[i] != want[i] {
			t.Errorf("got tags=%s, want tags=%s", got, want)
			return
		}
	}

	// pass nil val
	got = reflectutil.GetTagsFromTags(reflectutil.GetTagsFromTagsRequest{})
	if len(got) != 0 {
		t.Fatalf("pass nil value, got tags=%v, want tags=nil", got)
	}
}
