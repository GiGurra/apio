package apio

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type EndpointOutputBase interface {
	GetCode() int
	GetHeaders() map[string][]string
	GetBody() ([]byte, error)
}

func (e EndpointOutput[HeadersType, BodyType]) GetCode() int {
	return e.Code
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
