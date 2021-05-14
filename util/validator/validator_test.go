package validator_test

import (
	"sort"
	"testing"

	"github.com/pthethanh/micro/util/validator"
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
			got := validator.TagsToFieldNames(c.value, "json", c.tags...)
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

func TestValidatePartial(t *testing.T) {
	type Note struct {
		Value string `json:"value" validate:"required"`
	}
	type Address struct {
		Work string `json:"work" validate:"required"`
		Home string `json:"home" validate:"required"`
		Note Note   `json:"note"`
	}
	type Employee struct {
		Name     string  `json:"name"`
		Age      int     `json:"age" validate:"gt=1"`
		Address1 Address `json:"address1"`
		Note     string  `validate:"len=10"`
	}
	v := Employee{}
	if err := validator.ValidatePartial(v, "Name"); err != nil {
		t.Errorf("not required should not error, got err=%v, want err=nil", err)
	}
	if err := validator.ValidatePartial(v, "Age"); err == nil {
		t.Error("required shold have error, got err=nil, want err!=nil")
	}
	if err := validator.ValidatePartial(v, validator.TagsToFieldNames(v, "json", "address1.note.value")...); err == nil {
		t.Error("required nested fields extraction by TagsToStructFields, got err=nil, want err!=nil")
	}
}
