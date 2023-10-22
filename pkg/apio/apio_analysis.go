package apio

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

var cache = sync.Map{}

type FieldInfo struct {
	Name         string
	LKName       string
	FieldName    string
	OverrideName string
	StructField  reflect.StructField
	Type         reflect.Type
	ValueType    reflect.Type
	Index        int
	IsPointer    bool
}

func (a *FieldInfo) HasFieldNameInStruct() bool {
	return a.FieldName != ""
}

func (a *FieldInfo) String() string {
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

func (a *FieldInfo) IsRequired() bool {
	return !a.IsPointer
}

func (a *FieldInfo) IsOptional() bool {
	return !a.IsRequired()
}

func (a *FieldInfo) IsSlice() bool {
	return a.Type.Kind() == reflect.Slice
}

func (a *FieldInfo) Assign(parentPtr any, valuePtr any) error {
	parentT := reflect.TypeOf(parentPtr)
	if parentT.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer, got %v", parentT.Kind())
	}
	if parentT.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", parentT.Elem().Kind())
	}
	target := reflect.ValueOf(parentPtr).Elem().Field(a.Index)
	if a.IsPointer {
		target.Set(reflect.ValueOf(valuePtr))
	} else {
		if valuePtr == nil {
			return fmt.Errorf("expected non-nil value for required field %v", a.Name)
		}
		target.Set(reflect.ValueOf(valuePtr).Elem())
	}
	return nil
}

func (a *FieldInfo) GetPtr(parentPtr any) (any, error) {
	parentT := reflect.TypeOf(parentPtr)
	if parentT.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("expected pointer, got %v", parentT.Kind())
	}
	if parentT.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v", parentT.Elem().Kind())
	}
	source := reflect.ValueOf(parentPtr).Elem().Field(a.Index)
	if a.IsPointer {
		if source.IsNil() {
			return nil, nil
		} else {
			return source.Interface().(any), nil
		}
	} else {
		return source.Addr().Interface().(any), nil
	}
}

type StructInfo struct {
	Name              string
	Pkg               string
	Fields            []FieldInfo
	FieldsByFieldName map[string]FieldInfo
	FieldsByName      map[string]FieldInfo
	FieldsByLKName    map[string]FieldInfo // lower kebab case
}

func (a *StructInfo) HasContent() bool {
	return len(a.Fields) > 0
}

func GetFieldInfo[Parent any](parent Parent, index int) (FieldInfo, error) {
	return GetFieldInfoOfType(reflect.TypeOf(parent), index)
}

func GetFieldInfoOfType(parentT reflect.Type, index int) (FieldInfo, error) {
	structField := parentT.Field(index)
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

	return FieldInfo{
		Name:         name,
		LKName:       lkName,
		FieldName:    structField.Name,
		OverrideName: OverrideName,
		StructField:  structField,
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
			if i > 0 && !unicode.IsUpper(prev) && unicode.ToUpper(prev) != prev {
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

func GetStructInfo[T any](t T) (StructInfo, error) {
	return GetStructInfoOfType(reflect.TypeOf(t))
}
func GetStructInfoOfType(tpe reflect.Type) (StructInfo, error) {

	// Check that it is a struct
	if tpe.Kind() != reflect.Struct {
		return StructInfo{}, fmt.Errorf("expected struct, got %v", tpe.Kind())
	}

	structPkg := tpe.PkgPath()
	structName := tpe.Name()
	fullPath := fmt.Sprintf("%v/%v", structPkg, structName)

	cached, isCached := cache.Load(fullPath)
	if isCached {
		return cached.(StructInfo), nil
	}

	fields := make([]FieldInfo, 0)
	fieldsByFieldName := make(map[string]FieldInfo)
	fieldsByName := make(map[string]FieldInfo)
	fieldsByLKName := make(map[string]FieldInfo)

	for i := 0; i < tpe.NumField(); i++ {
		field := tpe.Field(i)

		analyzed, err := GetFieldInfoOfType(tpe, i)
		if err != nil {
			return StructInfo{}, fmt.Errorf("failed to analyze field %v: %v", field.Name, err)
		}
		fields = append(fields, analyzed)

		if analyzed.FieldName != "" && analyzed.FieldName != "_" {
			fieldsByFieldName[analyzed.FieldName] = analyzed
		}
		if analyzed.Name != "" && analyzed.Name != "_" {
			fieldsByName[analyzed.Name] = analyzed
		}
		if _, ok := fieldsByLKName[analyzed.LKName]; ok {
			return StructInfo{}, fmt.Errorf("duplicate lowercase field name %v", analyzed.Name)
		}
		if analyzed.LKName != "" && analyzed.LKName != "_" {
			fieldsByLKName[analyzed.LKName] = analyzed
		}
	}

	analyzed := StructInfo{
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
