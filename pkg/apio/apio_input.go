package apio

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	validateBodyType()
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

	var result EndpointInput[HeadersType, PathType, QueryType, BodyType]

	// parse headers
	for name, setter := range headerBindings.Bindings {
		inputValue := payload.Headers[name]
		rootStructValue := reflect.ValueOf(&result.Headers).Elem()
		valueToSet := rootStructValue.FieldByName(name)
		reflectZero := reflect.Value{}
		if valueToSet == reflectZero {
			// Did not find the correct field. Try to find it by tag
			rootStructType := rootStructValue.Type()
			for i := 0; i < rootStructType.NumField(); i++ {
				field := rootStructType.Field(i)
				if strings.ToLower(field.Tag.Get("name")) == strings.ToLower(name) {
					valueToSet = rootStructValue.Field(i)
					break
				}
			}
		}

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

	// parse body
	bodyT := reflect.TypeOf(e.Body)
	if bodyT.Kind() != reflect.Struct {
		return result, NewError(http.StatusInternalServerError, "BodyType must be a struct, this should have been caught in initial validation step", nil)
	}

	// num fields in body
	numFields := bodyT.NumField()
	if numFields >= 1 {
		err := json.Unmarshal(payload.Body, &result.Body)
		if err != nil {
			return result, fmt.Errorf("failed to unmarshal body: %w", err)
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

	// Iterate over fields in HeaderType
	for i := 0; i < pathT.NumField(); i++ {
		// Check if the field has path
		field := pathT.Field(i)

		if field.Name != "_" {
			name := field.Name
			if nameOvrd, ok := field.Tag.Lookup("name"); ok {
				name = nameOvrd
			}

			if alreadyTaken[name] {
				panic(fmt.Sprintf("header '%s' is already taken", name))
			}

			alreadyTaken[name] = true
			result.Bindings[name] = getFromStringHeaderFieldSetter(field, name)
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

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) validateBodyType() {
	bodyT := reflect.TypeOf(e.Body)
	if bodyT.Kind() != reflect.Struct {
		panic("BodyType must be a struct")
	}
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
