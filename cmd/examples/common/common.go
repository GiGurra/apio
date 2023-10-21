package common

import (
	. "github.com/GiGurra/apio/pkg/apio"
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

var GetEndpointSpec = Endpoint[
	EndpointInput[UserSettingHeaders, UserSettingPath, UserSettingQuery, X],
	EndpointOutput[X, UserSetting],
]{Method: http.MethodGet}

var PutEndpointSpec = Endpoint[
	EndpointInput[X, UserSettingPath, X, UserSetting],
	EndpointOutput[OutputHeaders, X],
]{Method: http.MethodPost}
