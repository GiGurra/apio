package apio

import (
	"fmt"
	"testing"
)

type TestStruct struct {
	RequiredField string
	OptionalField *int `name:"optional_field"`
}

func TestAnalyzeStruct(t *testing.T) {
	instance := TestStruct{}
	_, _ = AnalyzeStruct(instance)
	analyzed, err := AnalyzeStruct(instance)
	if err != nil {
		t.Fatalf("AnalyzeStruct returned an error: %v", err)
	}
	//
	//expected := AnalyzedStruct{
	//	Name: "TestStruct",
	//
	//}

	fmt.Printf("Analyzed struct: %+v\n", analyzed)
}
