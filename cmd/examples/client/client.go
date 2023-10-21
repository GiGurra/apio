package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/GiGurra/apio/cmd/examples/common"
	. "github.com/GiGurra/apio/pkg/apio"
	"io"
	"net/http"
)

func ptr[T any](v T) *T {
	return &v
}

func main() {
	server := Server{
		Scheme:   "http",
		Host:     "localhost",
		Port:     8080,
		BasePath: "/api/v1",
		HttpVer:  "1.1",
	}

	input := EndpointInput[UserSettingHeaders, UserSettingPath, UserSettingQuery, X]{
		Headers: UserSettingHeaders{
			Yo:          "yo",
			ContentType: "application/json",
		},
		Path: UserSettingPath{
			User:       123,
			SettingCat: "cat",
			SettingId:  "id",
		},
		Query: UserSettingQuery{
			Foo: ptr("foo"),
			Bar: 123,
		},
		Body: X{},
	}

	res, err := call(
		server,
		GetEndpointSpec,
		input,
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("res: %+v\n", res)
}

func call[Input EndpointInputBase, Output EndpointOutputBase](
	server Server,
	endpointSpec Endpoint[Input, Output],
	input Input,
) (Output, error) {

	var result Output
	payload, err := input.ToPayload()
	if err != nil {
		return result, fmt.Errorf("failed to convert input to payload: %w", err)
	}

	// Make http call
	client := http.Client{}
	bodyIoReader := bytes.NewReader(payload.Body)
	req, err := http.NewRequest(
		endpointSpec.Method,
		fmt.Sprintf("%s://%s:%d%s%s%s",
			server.Scheme,
			server.Host,
			server.Port,
			server.BasePath,
			payload.Path,
			payload.QueryString(),
		),
		bodyIoReader,
	)

	if err != nil {
		return result, fmt.Errorf("failed to create request: %w", err)
	}

	for k, vs := range payload.Headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("failed to make request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read response body: %w", err)
	}

	var outputBody Output
	err = json.Unmarshal(bodyBytes, &outputBody)
	if err != nil {
		return result, fmt.Errorf("failed to parse response body: %w", err)
	}

	newBodAny, err := result.SetBody(outputBody)
	if err != nil {
		return result, fmt.Errorf("failed to set body: %w", err)
	}
	result = newBodAny.(Output)

	newBodAny, err = result.SetHeaders(resp.Header)
	if err != nil {
		return result, fmt.Errorf("failed to set headers: %w", err)
	}
	result = newBodAny.(Output)

	newBodAny = result.SetCode(resp.StatusCode)
	result = newBodAny.(Output)

	return result, nil
}
