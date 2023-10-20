package apio

import "github.com/labstack/echo/v4"

func EchoInstall(echoServer *echo.Echo, api *Api) {
	for _, endpoint := range api.Endpoints {
		echoServer.Add(endpoint.GetMethod(), endpoint.GetPath(), func(ctx echo.Context) error {
			return nil
		})
	}
}
