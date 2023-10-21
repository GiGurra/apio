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
	SetBody(body any) (any, error)
	SetHeaders(hdrs map[string][]string) (any, error)
	SetCode(code int) any
}

func (e EndpointOutput[HeadersType, BodyType]) GetCode() int {
	return e.Code
}

func (e EndpointOutput[HeadersType, BodyType]) SetBody(nb any) (any, error) {
	bodyTyped, ok := nb.(BodyType)
	if !ok {
		var zero BodyType
		return e, fmt.Errorf("expected body to be of type %s, got %s", reflect.TypeOf(zero), reflect.TypeOf(nb))
	}
	e.Body = bodyTyped
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
