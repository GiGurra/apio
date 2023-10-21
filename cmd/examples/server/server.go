package main

import (
	"fmt"
	. "github.com/GiGurra/apio/cmd/examples/common"
	. "github.com/GiGurra/apio/pkg/apio"
	"github.com/labstack/echo/v4"
)

func UserSettingEndpoints() []EndpointBase {

	return []EndpointBase{

		GetEndpointSpec.
			WithHandler(func(
				input EndpointInput[UserSettingHeaders, UserSettingPath, UserSettingQuery, X],
			) (EndpointOutput[X, UserSetting], error) {
				fmt.Printf("invoked GET path with input: %+v\n", input)
				return BodyResponse(UserSetting{
					Value: "testValue",
					Type:  fmt.Sprintf("input=%+v", input),
				}), nil
			}),

		PutEndpointSpec.WithHandler(func(
			input EndpointInput[X, UserSettingPath, X, UserSetting],
		) (EndpointOutput[OutputHeaders, X], error) {
			fmt.Printf("invoked PUT path with input: %+v\n", input)
			return HeadersResponse(OutputHeaders{
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