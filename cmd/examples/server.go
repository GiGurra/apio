package main

import (
	"fmt"
	. "github.com/GiGurra/apio/pkg/apio"
	"github.com/labstack/echo/v4"
	"net/http"
)

// UserSettingPath represents "/users/:user/settings/:settingCat/:settingId"
type UserSettingPath struct {
	_          any `path:"/users"`
	User       int
	_          any `path:"/settings"`
	SettingCat string
	SettingId  string
}

type UserSettingQuery struct {
	Foo *string
	Bar int
}

type UserSetting struct {
	Value any     `json:"value"`
	Type  string  `json:"type"`
	Opt   *string `json:"opt"`
}

type UserSettingHeaders struct {
	Yo          any
	ContentType string `name:"Content-Type"`
}

type OutputHeaders struct {
	ContentType string `name:"Content-Type"`
}

func UserSettingEndpoints() []EndpointBase {

	return []EndpointBase{

		Endpoint[
			EndpointInput[UserSettingHeaders, UserSettingPath, UserSettingQuery, X],
			EndpointOutput[X, UserSetting],
		]{
			Method: http.MethodGet,
			Handler: func(
				input EndpointInput[UserSettingHeaders, UserSettingPath, UserSettingQuery, X],
			) (EndpointOutput[X, UserSetting], error) {
				fmt.Printf("invoked GET path with input: %+v\n", input)
				return BodyResponse(UserSetting{
					Value: "testValue",
					Type:  fmt.Sprintf("input=%+v", input),
				}), nil
			},
		},

		Endpoint[
			EndpointInput[X, UserSettingPath, X, UserSetting],
			EndpointOutput[OutputHeaders, X],
		]{
			Method: http.MethodPut,
			Handler: func(
				input EndpointInput[X, UserSettingPath, X, UserSetting],
			) (EndpointOutput[OutputHeaders, X], error) {
				fmt.Printf("invoked PUT path with input: %+v\n", input)
				return HeadersResponse(OutputHeaders{
					ContentType: "application/json",
				}), nil
			},
		},
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
