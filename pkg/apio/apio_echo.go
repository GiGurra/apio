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
				return endpoint.getPath()
			} else {
				return api.IntBasePath + "/" + strings.TrimPrefix(endpoint.getPath(), "/")
			}
		}()

		fmt.Printf(" * attaching endpoint: %s %s\n", endpoint.getMethod(), path)
		echoServer.Add(endpoint.getMethod(), path, func(ctx echo.Context) error {

			headers := map[string][]string{}
			for k, v := range ctx.Request().Header {
				headers[k] = v
			}

			path := map[string]string{}
			pathNames := ctx.ParamNames()
			pathValues := ctx.ParamValues()
			numIncParams := len(ctx.ParamNames())
			for i := 0; i < numIncParams; i++ {
				path[pathNames[i]] = pathValues[i]
			}

			query := map[string][]string{}
			for k, v := range ctx.QueryParams() {
				query[k] = v
			}

			body := ctx.Request().Body
			defer func(body io.ReadCloser) {
				err := body.Close()
				if err != nil {
					fmt.Printf("error closing body: %v", err)
				}
			}(body)

			bodyBytes, err := io.ReadAll(body)
			if err != nil {
				return fmt.Errorf("error reading body: %v", err)
			}

			result, err := endpoint.invoke(inputPayload{
				Headers: headers,
				Path:    path,
				Query:   query,
				Body:    bodyBytes,
			})

			if err != nil {
				var errResp *ErrResp
				if errors.As(err, &errResp) {
					fmt.Printf("error response: %v\n", errResp)
					return ctx.String(errResp.Code, errResp.ClMsg)
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

			return ctx.JSONBlob(result.GetCode(), outputBodyBytes)
		})
	}
}
