package apio

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"strings"
)

func EchoInstall(echoServer *echo.Echo, api *Api) {

	fmt.Printf("installing api '%s' to echo server: \n", api.Name)
	fmt.Printf(" * servers: \n")
	for _, s := range api.Servers {
		fmt.Printf("   * %+v: \n", s)
	}

	for i := range api.Endpoints {

		endpoint := api.Endpoints[i] // need to do this until go 1.22 is released

		path := func() string {
			if api.IntBasePath == "" {
				return endpoint.GetPathPattern()
			} else {
				return api.IntBasePath + "/" + strings.TrimPrefix(endpoint.GetPathPattern(), "/")
			}
		}()

		pathWithQueryParams := path + endpoint.GetQueryPattern()

		fmt.Printf(" * attaching endpoint: %s %s\n", endpoint.GetMethod(), pathWithQueryParams)
		echoServer.Add(endpoint.GetMethod(), path, func(ctx echo.Context) error {

			body := ctx.Request().Body
			defer func(body io.ReadCloser) {
				_, _ = io.ReadAll(body)
				err := body.Close()
				if err != nil {
					fmt.Printf("error closing body: %v", err)
				}
			}(body)

			headers := map[string][]string{}
			for k, v := range ctx.Request().Header {
				headers[k] = v
			}

			pathParams := map[string]string{}
			pathNames := ctx.ParamNames()
			pathValues := ctx.ParamValues()
			numIncParams := len(ctx.ParamNames())
			for i := 0; i < numIncParams; i++ {
				pathParams[pathNames[i]] = pathValues[i]
			}

			queryParams := map[string][]string{}
			for k, v := range ctx.QueryParams() {
				queryParams[k] = v
			}

			bodyBytes, err := io.ReadAll(body)
			if err != nil {
				return fmt.Errorf("error reading body: %v", err)
			}

			result, err := endpoint.Handle(InputPayload{
				Headers: headers,
				Path:    pathParams,
				Query:   queryParams,
				Body:    bodyBytes,
			})

			if err != nil {
				var errResp *ErrResp
				if errors.As(err, &errResp) {
					fmt.Printf("error response: %v\n", errResp)
					return ctx.String(errResp.Status, errResp.ClMsg)
				} else {
					fmt.Printf("error: %v\n", err)
					return ctx.String(500, fmt.Sprintf("internal error, see server logs"))
				}
			}

			outputBodyBytes, err := result.GetBody()
			if err != nil {
				fmt.Printf("error getting body: %v\n", err)
				return ctx.String(500, fmt.Sprintf("internal error, see server logs"))
			}

			// write headers
			for k, vs := range result.GetHeaders() {
				for _, v := range vs {
					ctx.Response().Header().Add(k, v)
				}
			}

			return ctx.JSONBlob(result.GetCode(), outputBodyBytes)
		})
	}
}
