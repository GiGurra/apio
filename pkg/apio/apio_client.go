package apio

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RPCOpts struct {
	Timeout time.Duration
}

func DefaultOpts() RPCOpts {
	return RPCOpts{
		Timeout: 15 * time.Second,
	}
}

func (e Endpoint[Input, Output]) RPC(
	server Server,
	input Input,
	opts RPCOpts,
) (Output, error) {

	var result Output
	payload, err := input.ToPayload()
	if err != nil {
		return result, fmt.Errorf("failed to convert input to payload: %w", err)
	}

	// Make http call
	client := http.Client{
		Timeout: opts.Timeout,
	}
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

	// Read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return result, ErrResp{
			Status: resp.StatusCode,
			ClMsg:  fmt.Sprintf("non-2xx status code: %d, body: %s", resp.StatusCode, string(bodyBytes)),
			IntErr: nil,
		}
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
