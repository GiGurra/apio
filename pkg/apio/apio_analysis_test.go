package apio

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"reflect"
	"testing"
)

type TestStruct struct {
	RequiredField string
	OptionalField *int `name:"optional_field"`
}

func TestAnalyzeStruct(t *testing.T) {
	instance := TestStruct{}
	analyzed, err := AnalyzeStruct(instance)
	if err != nil {
		t.Fatalf("AnalyzeStruct returned an error: %v", err)
	}

	expected := AnalyzedStruct{
		Name: "TestStruct",
		Pkg:  "github.com/GiGurra/apio/pkg/apio",
		Fields: []AnalyzedField{
			{
				Name:         "RequiredField",
				FieldName:    "RequiredField",
				LKName:       "required-field",
				OverrideName: "",
				Type:         reflect.TypeOf(""), // string
				ValueType:    reflect.TypeOf(""), // string
				Index:        0,
				IsPointer:    false,
			},
			{
				Name:         "optional_field",
				FieldName:    "OptionalField",
				LKName:       "optional_field",
				OverrideName: "optional_field",
				Type:         reflect.TypeOf(instance.OptionalField), // *int
				ValueType:    reflect.TypeOf(0),                      // int
				Index:        1,
				IsPointer:    true,
			},
		},
	}

	if expected.Name != analyzed.Name {
		t.Errorf("expected.Name != analyzed.Name: %v != %v", expected.Name, analyzed.Name)
	}

	if expected.Pkg != analyzed.Pkg {
		t.Fatalf("expected.Pkg != analyzed.Pkg: %v != %v", expected.Pkg, analyzed.Pkg)
	}

	for i := range expected.Fields {
		expField := expected.Fields[i]
		actField := analyzed.Fields[i]
		if diff := cmp.Diff(expField.String(), actField.String()); diff != "" {
			t.Fatalf("AnalyzedField mismatch:\n%s", diff)
		}
	}

	fmt.Printf("Analyzed struct: %+v\n", analyzed)
}
