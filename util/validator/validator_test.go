package validator_test

import (
	"testing"

	"github.com/pthethanh/micro/util/validator"
)

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
	if err := validator.ValidatePartial(v, "Address1.Note.Value"); err == nil {
		t.Error("required nested fields, got err=nil, want err!=nil")
	}
}
