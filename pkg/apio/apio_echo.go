package apio

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

func EchoInstall(echoServer *echo.Echo, api *Api) {

	slog.Info(fmt.Sprintf("installing api '%s' to echo server:", api.Name))
	slog.Info(fmt.Sprintf(" * servers: "))
	for _, s := range api.Servers {
		slog.Info(fmt.Sprintf("   * %+v", s))
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

		slog.Info(fmt.Sprintf(" * attaching endpoint: %s %s", endpoint.GetMethod(), pathWithQueryParams))
		echoServer.Add(endpoint.GetMethod(), path, func(ctx echo.Context) error {

			body := ctx.Request().Body
			defer func(body io.ReadCloser) {
				_, _ = io.ReadAll(body)
				err := body.Close()
				if err != nil {
					slog.Error(fmt.Sprintf("error closing body: %v", err))
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
					if errResp.Status/100 == 4 {
						slog.Warn(fmt.Sprintf("error response: %v", errResp))
					} else {
						slog.Error(fmt.Sprintf("error response: %v", errResp))
					}
					return ctx.String(errResp.Status, errResp.ClMsg)
				} else {
					slog.Error(fmt.Sprintf("error: %v", err))
					return ctx.String(500, fmt.Sprintf("internal error, see server logs"))
				}
			}

			outputBodyBytes, err := result.GetBody()
			if err != nil {
				slog.Error(fmt.Sprintf("error getting body: %v", err))
				return ctx.String(500, fmt.Sprintf("internal error, see server logs"))
			}

			// write headers
			for k, vs := range result.GetHeaders() {
				for _, v := range vs {
					ctx.Response().Header().Add(k, v)
				}
			}

			if len(outputBodyBytes) == 0 {
				return ctx.NoContent(http.StatusNoContent)
			} else {
				return ctx.JSONBlob(http.StatusOK, outputBodyBytes)
			}
		})
	}
}
