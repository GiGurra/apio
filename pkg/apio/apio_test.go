package apio

import (
	"github.com/labstack/echo/v4"
	"testing"
)

// represents "/users/:user/settings/:settingCat/:settingId"
type UserSettingPath struct {
	_          any `path:"/users"`
	user       int
	_          any `path:"/settings"`
	settingCat string
	settingId  string
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
			Method: "GET",
			Handler: func(
				input EndpointInput[X, UserSettingPath, X, X],
			) (EndpointOutput[X, UserSetting], error) {
				return BodyResponse(UserSetting{
					Value: "testValue",
					Type:  "testType",
				}), nil
			},
		},

		Endpoint[
			EndpointInput[X, UserSettingPath, X, UserSetting],
			EndpointOutput[X, X],
		]{
			Method: "PUT",
			Handler: func(
				input EndpointInput[X, UserSettingPath, X, UserSetting],
			) (EndpointOutput[X, X], error) {
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
		}.AddEndpoints(
			UserSettingEndpoints()...,
		)

	server := echo.New()

	EchoInstall(server, &api)
	//
	//err := server.Start(":8080")
	//if err != nil {
	//	t.Fatal(fmt.Errorf("failed to start server: %w", err))
	//}
}
