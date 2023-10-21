package apio

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func (e Endpoint[Input, Output]) Call(
	server Server,
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
	fullPath := fmt.Sprintf("%s://%s:%d%s%s%s",
		server.Scheme,
		server.Host,
		server.Port,
		server.BasePath,
		payload.PathStr,
		payload.QueryString(),
	)
	req, err := http.NewRequest(
		e.Method,
		fullPath,
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

	if resp.StatusCode/100 != 2 {
		return result, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	// Read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read response body: %w", err)
	}

	newBodAny, err := result.SetBody(bodyBytes)
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
