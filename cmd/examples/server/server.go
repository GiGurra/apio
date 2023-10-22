package main

import (
	"encoding/json"
	"fmt"
	"github.com/GiGurra/apio/cmd/examples/user_setting"
	. "github.com/GiGurra/apio/pkg/apio"
	"github.com/GiGurra/apio/pkg/apio/openapi3"
	"github.com/labstack/echo/v4"
)

func UserSettingEndpoints() []EndpointBase {

	return []EndpointBase{

		user_setting.GetAll.
			WithHandler(func(
				input EndpointInput[X, user_setting.PathAll, user_setting.Query, X],
			) (EndpointOutput[user_setting.RespHeaders, []user_setting.Body], error) {
				fmt.Printf("invoked GET path with input: %+v\n", input)
				return Response(
					user_setting.RespHeaders{
						ContentType: "application/json",
					},
					[]user_setting.Body{{
						Value: "testValue",
						Type:  fmt.Sprintf("input=%+v", input),
					}},
				), nil
			}),

		user_setting.GetById.
			WithHandler(func(
				input EndpointInput[X, user_setting.PathById, X, X],
			) (EndpointOutput[user_setting.RespHeaders, user_setting.Body], error) {
				fmt.Printf("invoked GET path with input: %+v\n", input)
				return Response(
					user_setting.RespHeaders{
						ContentType: "application/json",
					},
					user_setting.Body{
						Value: "testValue",
						Type:  fmt.Sprintf("input=%+v", input),
					},
				), nil
			}),

		user_setting.PutById.
			WithHandler(func(
				input EndpointInput[user_setting.Headers, user_setting.PathById, X, user_setting.Body],
			) (EndpointOutput[X, X], error) {
				fmt.Printf("invoked PUT path with input: %+v\n", input)
				return EmptyResponse(), nil
			}),
	}
}

func main() {
	testApi := Api{
		Name:        "My test API",
		Description: "This is a test API.",
		Version:     "1.0.0",
		Servers: []Server{{
			Name:        "My test server",
			Description: "My test server description",
			Scheme:      "https",
			Host:        "api.example.com",
			Port:        443,
			BasePath:    "/api/v1",
			HttpVer:     "1.1",
		}},
		IntBasePath: "/api/v1",
	}.WithEndpoints(
		UserSettingEndpoints()...,
	).Validate(true)

	openApi3 := openapi3.ToOpenApi3(testApi)
	openApi3Json, err := json.MarshalIndent(openApi3, "", "  ")
	if err != nil {
		panic(fmt.Errorf("failed to marshal OpenAPI 3 spec: %v", err))
	}

	fmt.Printf("OpenAPI 3 spec:\n")
	fmt.Printf("%s\n", openApi3Json)

	echoServer := echo.New()

	// You can use whatever router/server you want, I just happened to use Echo here.
	// apio comes with a helper function to install the endpoints on an Echo server,
	// but you can also do this manually/add your own integration if you want to.
	// see https://github.com/GiGurra/apio/blob/main/pkg/apio/apio_echo.go
	EchoInstall(echoServer, &testApi)

	_ = echoServer.Start(":8080")
}
