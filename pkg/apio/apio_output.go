package apio

import (
	"encoding/json"
	"fmt"
)

type EndpointOutputBase interface {
	GetCode() int
	GetHeaders() map[string][]string
	GetBody() ([]byte, error)
}

type EndpointOutput[
	HeadersType any,
	BodyType any,
] struct {
	Code    int
	Headers HeadersType
	Body    BodyType
}

func (e EndpointOutput[HeadersType, BodyType]) GetCode() int {
	return e.Code
}

func (e EndpointOutput[HeadersType, BodyType]) GetHeaders() map[string][]string {
	fmt.Printf("TODO: implement EndpointOutput.GetHeaders\n")
	return make(map[string][]string)
}

func (e EndpointOutput[HeadersType, BodyType]) GetBody() ([]byte, error) {
	bytes, err := json.Marshal(e.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}
	return bytes, nil
}
