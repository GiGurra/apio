package apio

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"testing"
)

// represents "/users/:user/settings/:settingCat/:settingId"
type UserSettingPath struct {
	_          any `path:"/users"`
	User       int
	_          any `path:"/settings"`
	SettingCat string
	SettingId  string
}

type UserSetting struct {
	Value any    `json:"value"`
	Type  string `json:"type"`
}

func UserSettingEndpoints() []EndpointBase {

	return []EndpointBase{

		Endpoint[
			EndpointInput[X, UserSettingPath, X, X],
			EndpointOutput[X, UserSetting],
		]{
			Method: http.MethodGet,
			Handler: func(
				input EndpointInput[X, UserSettingPath, X, X],
			) (EndpointOutput[X, UserSetting], error) {
				fmt.Printf("invoked GET path with input: %+v\n", input)
				return BodyResponse(UserSetting{
					Value: "testValue",
					Type:  fmt.Sprintf("input=%+v", input.Path),
				}), nil
			},
		},

		Endpoint[
			EndpointInput[X, UserSettingPath, X, UserSetting],
			EndpointOutput[X, X],
		]{
			Method: http.MethodPut,
			Handler: func(
				input EndpointInput[X, UserSettingPath, X, UserSetting],
			) (EndpointOutput[X, X], error) {
				fmt.Printf("invoked PUT path with input: %+v\n", input)
				return EmptyResponse(), nil
			},
		},
	}
}

func TestGetUserSetting(t *testing.T) {

	api :=
		Api{
			Published: []Server{{
				Scheme:   "https",
				Host:     "api.example.com",
				Port:     443,
				BasePath: "/api/v1",
			}},
			IntBasePath: "/api/v1",
		}.WithEndpoints(
			UserSettingEndpoints()...,
		)

	server := echo.New()

	// add recovery middleware
	//server.Use(middleware.Recover())

	fmt.Printf("api: %+v\n", api)

	EchoInstall(server, &api)

	err := server.Start(":8080")
	if err != nil {
		t.Fatal(fmt.Errorf("failed to start server: %w", err))
	}
}
