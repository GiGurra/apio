package apio

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

var cache = sync.Map{}

type AnalyzedField struct {
	Name         string
	LKName       string
	FieldName    string
	OverrideName string
	Type         reflect.Type
	ValueType    reflect.Type
	Index        int
	IsPointer    bool
}

func (a *AnalyzedField) String() string {
	result := a.Name + "{\n"
	result += fmt.Sprintf(" Name: %v\n", a.Name)
	result += fmt.Sprintf(" LKName: %v\n", a.LKName)
	result += fmt.Sprintf(" FieldName: %v\n", a.FieldName)
	result += fmt.Sprintf(" OverrideName: %v\n", a.OverrideName)
	result += fmt.Sprintf(" Type: %v\n", a.Type)
	result += fmt.Sprintf(" ValueType: %v\n", a.ValueType)
	result += fmt.Sprintf(" Index: %v\n", a.Index)
	result += fmt.Sprintf(" IsPointer: %v\n", a.IsPointer)
	result += "}"
	return result
}

func (a *AnalyzedField) IsRequired() bool {
	return !a.IsPointer
}

func (a *AnalyzedField) IsOptional() bool {
	return !a.IsRequired()
}

func (a *AnalyzedField) IsSlice() bool {
	return a.Type.Kind() == reflect.Slice
}

func (a *AnalyzedField) Assign(parentStruct *any, value *any) error {
	parentT := reflect.TypeOf(parentStruct)
	if parentT.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", parentT.Kind())
	}
	target := reflect.ValueOf(parentStruct).Elem().Field(a.Index)
	if a.IsPointer {
		target.Set(reflect.ValueOf(value))
	} else {
		if value == nil {
			return fmt.Errorf("expected non-nil value for required field %v", a.Name)
		}
		target.Set(reflect.ValueOf(value).Elem())
	}
	return nil
}

type AnalyzedStruct struct {
	Name              string
	Pkg               string
	Fields            []AnalyzedField
	FieldsByFieldName map[string]AnalyzedField
	FieldsByName      map[string]AnalyzedField
	FieldsByLKName    map[string]AnalyzedField // lower kebab case
}

func AnalyzeField[Parent any](parent Parent, index int) (AnalyzedField, error) {
	structField := reflect.TypeOf(parent).Field(index)
	fieldType := structField.Type
	isPointer := fieldType.Kind() == reflect.Ptr

	name := structField.Name
	OverrideName := ""
	if tag, ok := structField.Tag.Lookup("name"); ok {
		name = tag
		OverrideName = tag
	}
	lkName := camelCaseToKebabCase(name)

	valueType := fieldType
	if fieldType.Kind() == reflect.Ptr {
		valueType = fieldType.Elem()
	}

	return AnalyzedField{
		Name:         name,
		LKName:       lkName,
		FieldName:    structField.Name,
		OverrideName: OverrideName,
		Type:         fieldType,
		ValueType:    valueType,
		Index:        index,
		IsPointer:    isPointer,
	}, nil
}

func camelCaseToKebabCase(in string) string {
	out := ""
	var prev rune
	for i, c := range in {
		// Using unicode/runes
		if unicode.IsUpper(c) {
			if i > 0 && !unicode.IsUpper(prev) {
				out += "-"
			}
			out += string(unicode.ToLower(c))
		} else {
			out += string(c)
		}
		prev = c
	}
	return strings.ToLower(out)
}

func AnalyzeStruct[T any](t T) (AnalyzedStruct, error) {

	// Check that it is a struct
	tpe := reflect.TypeOf(t)
	if tpe.Kind() != reflect.Struct {
		return AnalyzedStruct{}, fmt.Errorf("expected struct, got %v", tpe.Kind())
	}

	structPkg := tpe.PkgPath()
	structName := tpe.Name()
	fullPath := fmt.Sprintf("%v/%v", structPkg, structName)

	cached, isCached := cache.Load(fullPath)
	if isCached {
		return cached.(AnalyzedStruct), nil
	}

	fields := make([]AnalyzedField, tpe.NumField())
	fieldsByFieldName := make(map[string]AnalyzedField)
	fieldsByName := make(map[string]AnalyzedField)
	fieldsByLKName := make(map[string]AnalyzedField)

	for i := 0; i < tpe.NumField(); i++ {
		field := tpe.Field(i)
		analyzed, err := AnalyzeField(t, i)
		if err != nil {
			return AnalyzedStruct{}, fmt.Errorf("failed to analyze field %v: %v", field.Name, err)
		}
		fields[i] = analyzed
		fieldsByFieldName[analyzed.FieldName] = analyzed
		fieldsByName[analyzed.Name] = analyzed
		if _, ok := fieldsByLKName[analyzed.LKName]; ok {
			return AnalyzedStruct{}, fmt.Errorf("duplicate lowercase field name %v", analyzed.Name)
		}
		fieldsByLKName[analyzed.LKName] = analyzed
	}

	analyzed := AnalyzedStruct{
		Name:              structName,
		Pkg:               structPkg,
		Fields:            fields,
		FieldsByFieldName: fieldsByFieldName,
		FieldsByName:      fieldsByName,
		FieldsByLKName:    fieldsByLKName,
	}
	cache.Store(fullPath, analyzed)
	return analyzed, nil
}
