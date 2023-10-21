package apio

import (
	"fmt"
	"reflect"
	"strings"
)

type inputPayload struct {
	Headers map[string][]string
	Path    map[string]string
	Query   map[string][]string
	Body    []byte
}

type endpointInputBase interface {
	getHeaders() any
	getPath() any
	calcHeaderBindings() HeaderBindings
	calcPathBindings() PathBindings
	calcQueryBindings() QueryBindings
	getQuery() any
	getBody() any
	parse(
		payload inputPayload,
		bindings HeaderBindings,
		pathBinding PathBindings,
		queryBindings QueryBindings,
	) (any, error)
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) getHeaders() any {
	return e.Headers
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) getPath() any {
	return e.Path
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) parse(
	payload inputPayload,
	headerBindings HeaderBindings,
	pathBindings PathBindings,
	queryBindings QueryBindings,
) (any, error) {
	fmt.Printf("todo: implement EndpointInput.parse\n")

	var result EndpointInput[HeadersType, PathType, QueryType, BodyType]

	// parse headers
	for name, setter := range headerBindings.Bindings {
		inputValue := payload.Headers[name]
		valueToSet := reflect.ValueOf(&result.Headers).Elem().FieldByName(name)
		if len(inputValue) > 1 {
			return result, fmt.Errorf("repeated header parameters not yet supported, field: %s", name)
		}
		var err error
		if len(inputValue) == 0 {
			err = setter(valueToSet, nil)
		} else {
			err = setter(valueToSet, &inputValue[0])
		}
		if err != nil {
			return result, fmt.Errorf("failed to set header parameter '%s': %w", name, err)
		}
	}

	// parse path parameters
	for name, setter := range pathBindings.Bindings {
		inputValue, ok := payload.Path[name]
		if !ok {
			return result, fmt.Errorf("missing path parameter '%s'", name)
		}
		valueToSet := reflect.ValueOf(&result.Path).Elem().FieldByName(name)
		err := setter(valueToSet, inputValue)
		if err != nil {
			return result, fmt.Errorf("failed to set path parameter '%s': %w", name, err)
		}
	}

	// parse query parameters
	for name, setter := range queryBindings.Bindings {
		inputValue := payload.Query[name]
		valueToSet := reflect.ValueOf(&result.Query).Elem().FieldByName(name)
		if len(inputValue) > 1 {
			return result, fmt.Errorf("repeated query parameters not yet supported, field: %s", name)
		}
		var err error
		if len(inputValue) == 0 {
			err = setter(valueToSet, nil)
		} else {
			err = setter(valueToSet, &inputValue[0])
		}
		if err != nil {
			return result, fmt.Errorf("failed to set query parameter '%s': %w", name, err)
		}
	}
	return result, nil
}

type HeaderBindings struct {
	Bindings map[string]headerFieldSetter
}

type PathBindings struct {
	FlatPath string
	Bindings map[string]pathFieldSetter
}

type QueryBindings struct {
	FlatPath string
	Bindings map[string]queryFieldSetter
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) calcHeaderBindings() HeaderBindings {

	pathT := reflect.TypeOf((*HeadersType)(nil)).Elem()
	if pathT.Kind() != reflect.Struct {
		panic("HeadersType must be a struct")
	}

	result := HeaderBindings{
		Bindings: make(map[string]headerFieldSetter),
	}
	alreadyTaken := make(map[string]bool)

	// Iterate over fields in PathType
	for i := 0; i < pathT.NumField(); i++ {
		// Check if the field has path
		field := pathT.Field(i)

		if field.Name != "_" {
			if alreadyTaken[field.Name] {
				panic(fmt.Sprintf("field '%s' is already taken", field.Name))
			}

			alreadyTaken[field.Name] = true
			result.Bindings[field.Name] = getFromStringQueryFieldSetter(field)
		}
	}

	return result
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) calcPathBindings() PathBindings {

	pathT := reflect.TypeOf((*PathType)(nil)).Elem()
	if pathT.Kind() != reflect.Struct {
		panic(fmt.Sprintf("PathType must be a struct, but is a %s", pathT.Kind().String()))
	}

	result := PathBindings{
		FlatPath: "",
		Bindings: make(map[string]pathFieldSetter),
	}
	alreadyTaken := make(map[string]bool)

	// Iterate over fields in PathType
	for i := 0; i < pathT.NumField(); i++ {
		// Check if the field has path
		field := pathT.Field(i)

		if field.Name == "_" {
			// We won't bind this parameter, but it is still needed in the path
			// Check if it has a tag called path
			pathTag := field.Tag.Get("path")
			if pathTag == "" {
				// Treat as wildcard
				result.FlatPath += "/*"
			} else {
				// Treat as literal
				result.FlatPath += "/" + strings.TrimPrefix(pathTag, "/")
			}
		} else {
			if alreadyTaken[field.Name] {
				panic(fmt.Sprintf("field '%s' is already taken", field.Name))
			}

			alreadyTaken[field.Name] = true
			result.FlatPath += "/:" + field.Name
			result.Bindings[field.Name] = getFromStringPathFieldSetter(field)
		}
	}

	return result
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) calcQueryBindings() QueryBindings {

	pathT := reflect.TypeOf((*QueryType)(nil)).Elem()
	if pathT.Kind() != reflect.Struct {
		panic("QueryType must be a struct")
	}

	result := QueryBindings{
		Bindings: make(map[string]queryFieldSetter),
	}
	alreadyTaken := make(map[string]bool)

	// Iterate over fields in PathType
	for i := 0; i < pathT.NumField(); i++ {
		// Check if the field has path
		field := pathT.Field(i)

		if field.Name != "_" {
			if alreadyTaken[field.Name] {
				panic(fmt.Sprintf("field '%s' is already taken", field.Name))
			}

			isFirst := len(result.Bindings) == 0
			separator := "&"
			if isFirst {
				separator = "?"
			}
			alreadyTaken[field.Name] = true
			result.FlatPath += separator + field.Name + "=.."
			result.Bindings[field.Name] = getFromStringQueryFieldSetter(field)
		}
	}

	return result
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) getQuery() any {
	return e.Query
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) getBody() any {
	return e.Body
}

func calcHeaderBindings[Input endpointInputBase]() HeaderBindings {
	var zero Input
	return zero.calcHeaderBindings()
}

func calcPathBindings[Input endpointInputBase]() PathBindings {
	var zero Input
	return zero.calcPathBindings()
}

func calcQueryBindings[Input endpointInputBase]() QueryBindings {
	var zero Input
	return zero.calcQueryBindings()
}
