package main

import (
	"fmt"
	"github.com/GiGurra/apio/cmd/examples/user_setting"
	. "github.com/GiGurra/apio/pkg/apio"
	"github.com/labstack/echo/v4"
)

func UserSettingEndpoints() []EndpointBase {

	return []EndpointBase{

		user_setting.Get.
			WithHandler(func(
				input EndpointInput[user_setting.Headers, user_setting.Path, user_setting.Query, X],
			) (EndpointOutput[X, user_setting.Body], error) {
				fmt.Printf("invoked GET path with input: %+v\n", input)
				return BodyResponse(user_setting.Body{
					Value: "testValue",
					Type:  fmt.Sprintf("input=%+v", input),
				}), nil
			}),

		user_setting.Put.WithHandler(func(
			input EndpointInput[X, user_setting.Path, X, user_setting.Body],
		) (EndpointOutput[user_setting.RespHeaders, X], error) {
			fmt.Printf("invoked PUT path with input: %+v\n", input)
			return HeadersResponse(user_setting.RespHeaders{
				ContentType: "application/json",
			}), nil
		}),
	}
}

func main() {
	var testApi = Api{
		Name: "My test API",
		Servers: []Server{{
			Scheme:   "https",
			Host:     "api.example.com",
			Port:     443,
			BasePath: "/api/v1",
			HttpVer:  "1.1",
		}},
		IntBasePath: "/api/v1",
	}.WithEndpoints(
		UserSettingEndpoints()...,
	).Validate()

	echoServer := echo.New()

	EchoInstall(echoServer, &testApi)

	_ = echoServer.Start(":8080")
}
