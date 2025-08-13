package util

import (
	"errors"
	"testing"
)

type sample struct {
	Name string `yaml:"name" validate:"required"`
}

// TestUnmarshalYamlIntoStruct_OK verifies successful unmarshalling and validation.
func TestUnmarshalYamlIntoStruct_OK(t *testing.T) {
	var s sample
	y := "name: test"
	if err := UnmarshalYamlIntoStruct(y, &s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Name != "test" {
		t.Fatalf("got %q", s.Name)
	}
}

// TestUnmarshalYamlIntoStruct_Strict ensures unknown fields trigger a strict error.
func TestUnmarshalYamlIntoStruct_Strict(t *testing.T) {
	var s sample
	y := "name: test\nextra: nope"
	err := UnmarshalYamlIntoStruct(y, &s)
	if err == nil {
		t.Fatalf("expected strict error for unknown field")
	}
}

// TestUnmarshalYamlIntoStruct_Validation checks that required field validation fails.
func TestUnmarshalYamlIntoStruct_Validation(t *testing.T) {
	var s sample
	y := "{}"
	if err := UnmarshalYamlIntoStruct(y, &s); err == nil {
		t.Fatalf("expected validation error")
	} else if !errors.Is(err, err) { // just ensure non-nil
		// no-op
	}
}
