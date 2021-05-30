package reflectutil_test

import (
	"sort"
	"testing"

	"github.com/pthethanh/micro/util/reflectutil"
)

func TestTagsToStructFields(t *testing.T) {
	type Note struct {
		Value string `json:"value"`
	}
	type Address struct {
		Work string `json:"work"`
		Home string `json:"home"`
		Note Note   `json:"note"`
	}
	v := struct {
		Name     string   `json:"name"`
		Age      int      `json:"age"`
		Address1 Address  `json:"address1"`
		Address2 *Address `json:"address2"`
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
			name:  "all fields - nil",
			value: v,
			tags:  nil,
			want:  []string{"Name", "Age", "Address1.Note", "Address2", "Address2.Work", "Address2.Home", "Address2.Note.Value", "Address1", "Address1.Work", "Address1.Home", "Address1.Note.Value", "Address2.Note"},
		},
		{
			name:  "all fields - *",
			value: v,
			tags:  []string{"*"},
			want:  []string{"Name", "Age", "Address1.Note", "Address2", "Address2.Work", "Address2.Home", "Address2.Note.Value", "Address1", "Address1.Work", "Address1.Home", "Address1.Note.Value", "Address2.Note"},
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
			got := reflectutil.GetFieldNamesFromTags(c.value, "json", c.tags...)
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
		Value string `json:"value" bson:"bvalue"`
	}
	type Address struct {
		Work string `json:"work" bson:"bwork"`
		Home string `json:"home" bson:"bhome"`
		Note Note   `json:"note" bson:"bnote"`
	}
	v := struct {
		Name     string   `json:"name" bson:"bname"`
		Age      int      `json:"age" bson:"bage"`
		Address1 Address  `json:"address1" bson:"baddress1"`
		Address2 *Address `json:"address2" bson:"baddress2"`
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
			name:  "some fields different levels",
			value: v,
			src:   "json",
			dst:   "bson",
			tags:  []string{"name", "age", "address1.work", "address2.note.value"},
			want:  []string{"bname", "bage", "baddress1.bwork", "baddress2.bnote.bvalue"},
		},
		{
			name:  "all fields - nil",
			value: v,
			tags:  nil,
			want:  []string{"bname", "bage", "baddress1.bnote", "baddress2", "baddress2.bwork", "baddress2.bhome", "baddress2.bnote.bvalue", "baddress1", "baddress1.bwork", "baddress1.bhome", "baddress1.bnote.bvalue", "baddress2.bnote"},
		},
		{
			name:  "all fields - *",
			value: v,
			tags:  []string{"*"},
			want:  []string{"bname", "bage", "baddress1.bnote", "baddress2", "baddress2.bwork", "baddress2.bhome", "baddress2.bnote.bvalue", "baddress1", "baddress1.bwork", "baddress1.bhome", "baddress1.bnote.bvalue", "baddress2.bnote"},
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
			got := reflectutil.GetTagsFromTags(c.value, "json", "bson", c.tags...)
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
