package apio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type OutputPayload struct {
	Headers map[string][]string
	Body    []byte
}

type EndpointOutputBase interface {
	GetHeaders() map[string][]string
	GetBody() ([]byte, error)
	ToPayload() (OutputPayload, error)
	validateBodyType()
	validateHeadersType()
	SetBody(jsonBytes []byte) (EndpointOutputBase, error)
	SetHeaders(hdrs map[string][]string) (EndpointOutputBase, error)
	SetAll(hdrs map[string][]string, jsonBodyBytes []byte) (EndpointOutputBase, error)
	GetBodyInfo() StructInfo
	GetDescription() string
	OkCode() int
}

func (e EndpointOutput[HeadersType, BodyType]) GetDescription() string {
	return e.Description
}

func (e EndpointOutput[HeadersType, BodyType]) GetBodyInfo() StructInfo {
	info, err := GetStructInfo(e.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to analyze struct: %v", err))
	}
	return info
}

func (e EndpointOutput[HeadersType, BodyType]) OkCode() int {
	bodyInfo, err := GetStructInfo(e.Body)
	if err != nil {
		panic(fmt.Errorf("failed to analyze struct: %w", err))
	}
	if len(bodyInfo.Fields) == 0 {
		return http.StatusNoContent
	} else {
		return http.StatusOK
	}
}

func (e EndpointOutput[HeadersType, BodyType]) SetBody(jsonBytes []byte) (EndpointOutputBase, error) {

	// if target has no fields, just return
	if reflect.TypeOf(e.Body).NumField() == 0 {
		return e, nil
	}

	var res BodyType
	err := json.Unmarshal(jsonBytes, &res)
	if err != nil {
		return e, fmt.Errorf("failed to unmarshal body: %w", err)
	}
	e.Body = res
	return e, nil
}

func (e EndpointOutput[HeadersType, BodyType]) SetAll(hdrs map[string][]string, jsonBodyBytes []byte) (EndpointOutputBase, error) {
	res, err := e.SetHeaders(hdrs)
	if err != nil {
		return res, fmt.Errorf("failed to set headers: %w", err)
	}
	res, err = res.SetBody(jsonBodyBytes)
	if err != nil {
		return res, fmt.Errorf("failed to set body: %w", err)
	}
	return res, nil
}

func (e EndpointOutput[HeadersType, BodyType]) SetHeaders(hdrs map[string][]string) (EndpointOutputBase, error) {

	// Check that it is a struct
	structInfo, err := GetStructInfo(e.Headers)
	if err != nil {
		panic(fmt.Errorf("failed to analyze headers struct: %w", err))
	}

	requiredNotSet := make(map[string]bool)
	for _, field := range structInfo.Fields {
		if field.IsRequired() {
			requiredNotSet[field.LKName] = true
		}
	}

	// Check that all headers are present
	for k, vs := range hdrs {
		for _, v := range vs {
			lkName := strings.ToLower(k)
			field, exists := structInfo.FieldsByLKName[lkName]
			if !exists {
				continue // ignore extra headers
			}

			parser, err := getStringParsePtrFn(field.Type)
			if err != nil {
				return e, fmt.Errorf("failed to get parser for header '%s': %w", k, err)
			}
			newValuePtr, err := parser(v)
			if err != nil {
				return e, fmt.Errorf("failed to parse header '%s': %w", k, err)
			}
			// assign the value to the struct field
			// Check if it is a pointer first, in which case we need to set it using an address
			if reflect.ValueOf(e.Headers).Field(field.Index).Kind() == reflect.Ptr {
				reflect.ValueOf(&e.Headers).Elem().Field(field.Index).Set(reflect.ValueOf(newValuePtr))
			} else {
				reflect.ValueOf(&e.Headers).Elem().Field(field.Index).Set(reflect.ValueOf(newValuePtr).Elem())
			}
			delete(requiredNotSet, lkName)
		}
	}

	if len(requiredNotSet) > 0 {
		return e, fmt.Errorf("required headers not set: %v", requiredNotSet)
	}

	return e, nil
}

func (e EndpointOutput[HeadersType, BodyType]) ToPayload() (OutputPayload, error) {
	bodyBytes, err := e.GetBody()
	if err != nil {
		return OutputPayload{}, fmt.Errorf("failed to get body: %w", err)
	}
	return OutputPayload{
		Headers: e.GetHeaders(),
		Body:    bodyBytes,
	}, nil
}

func (e EndpointOutput[HeadersType, BodyType]) GetHeaders() map[string][]string {
	result := make(map[string][]string)

	// Check that it is a struct
	headersType := reflect.TypeOf(e.Headers)
	if headersType.Kind() != reflect.Struct {
		panic(fmt.Errorf("expected output headers to be a struct, got %s", headersType.Kind()))
	}

	numFields := headersType.NumField()
	for i := 0; i < numFields; i++ {
		field := headersType.Field(i)
		if field.Name != "_" {
			name := field.Name
			if nameOvrd, ok := field.Tag.Lookup("name"); ok {
				name = nameOvrd
			}
			result[name] = []string{fmt.Sprintf("%v", reflect.ValueOf(e.Headers).Field(i).Interface())}
		}
	}

	return result
}

func (e EndpointOutput[HeadersType, BodyType]) GetBody() ([]byte, error) {
	structInfo, err := GetStructInfo(e.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze struct: %w", err)
	}
	if len(structInfo.Fields) == 0 {
		return []byte{}, nil
	}
	bytes, err := json.Marshal(e.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}
	return bytes, nil
}

func (e EndpointOutput[HeadersType, BodyType]) validateBodyType() {
	bodyT := reflect.TypeOf(e.Body)
	if bodyT.Kind() != reflect.Struct {
		panic("BodyType must be a struct")
	}
}

func (e EndpointOutput[HeadersType, BodyType]) validateHeadersType() {
	bodyT := reflect.TypeOf(e.Headers)
	if bodyT.Kind() != reflect.Struct {
		panic("HeadersType must be a struct")
	}
}
