package apio

import (
	"fmt"
	"github.com/labstack/echo/v4"
)

func EchoInstall(echoServer *echo.Echo, api *Api) {
	for _, endpoint := range api.Endpoints {
		fmt.Printf("%s %s\n", endpoint.GetMethod(), endpoint.GetPath())
		//echoServer.Add(endpoint.GetMethod(), endpoint.GetPath(), func(ctx echo.Context) error {
		//	return nil
		//})
	}
}
