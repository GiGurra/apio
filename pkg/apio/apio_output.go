package apio

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type OutputPayload struct {
	Code    int
	Headers map[string][]string
	Body    []byte
}

type EndpointOutputBase interface {
	GetCode() int
	GetHeaders() map[string][]string
	GetBody() ([]byte, error)
	ToPayload() (OutputPayload, error)
	validateBodyType()
	validateHeadersType()
	SetBody(jsonBytes []byte) (any, error)
	SetHeaders(hdrs map[string][]string) (any, error)
	SetCode(code int) any
}

func (e EndpointOutput[HeadersType, BodyType]) GetCode() int {
	return e.Code
}

func (e EndpointOutput[HeadersType, BodyType]) SetBody(jsonBytes []byte) (any, error) {

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

func (e EndpointOutput[HeadersType, BodyType]) SetCode(code int) any {
	e.Code = code
	return e
}

func (e EndpointOutput[HeadersType, BodyType]) SetHeaders(hdrs map[string][]string) (any, error) {

	// Check that it is a struct
	headersType := reflect.TypeOf(e.Headers)
	if headersType.Kind() != reflect.Struct {
		panic(fmt.Errorf("expected output headers to be a struct, got %s", headersType.Kind()))
	}

	// Check that all headers are present
	fmt.Printf("WARNING: NOT YET implemented: Headers ignored\n")

	return e, nil
}

func (e EndpointOutput[HeadersType, BodyType]) ToPayload() (OutputPayload, error) {
	bodyBytes, err := e.GetBody()
	if err != nil {
		return OutputPayload{}, fmt.Errorf("failed to get body: %w", err)
	}
	return OutputPayload{
		Code:    e.Code,
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
