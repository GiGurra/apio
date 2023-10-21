package apio

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"testing"
)

func UserSettingEndpoints() []EndpointBase {

	// represents "/users/:collection/settings/:category/:value"
	type Path struct {
		_          any `path:"/users"`
		collection int
		_          any `path:"/settings"`
		category   string
		value      string
	}

	type UserSettingBody struct {
		Value any    `json:"value"`
		Type  string `json:"type"`
	}

	return []EndpointBase{

		Endpoint[
			EndpointInput[X, Path, X, X],
			EndpointOutput[X, UserSettingBody],
		]{}.WithMethod("GET").
			WithHandler(func(
				input EndpointInput[X, Path, X, X],
			) (EndpointOutput[X, UserSettingBody], *Error) {
				return BodyResponse(UserSettingBody{
					Value: "testValue",
					Type:  "testType",
				}), nil
			}),

		Endpoint[
			EndpointInput[X, Path, X, UserSettingBody],
			EndpointOutput[X, X],
		]{}.WithMethod("PUT").
			WithHandler(func(
				input EndpointInput[X, Path, X, UserSettingBody],
			) (EndpointOutput[X, X], *Error) {
				return EmptyResponse(), nil
			}),
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
		}.AddEndpoints(
			UserSettingEndpoints()...,
		)

	server := echo.New()

	EchoInstall(server, &api)

	err := server.Start(":8080")
	if err != nil {
		t.Fatal(fmt.Errorf("failed to start server: %w", err))
	}
}
