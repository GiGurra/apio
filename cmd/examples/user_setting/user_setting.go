package user_setting

import (
	. "github.com/GiGurra/apio/pkg/apio"
	"net/http"
)

// Path represents "/users/:user/settings/:settingCat/:settingId"
type Path struct {
	_          any `path:"/users"`
	User       int
	_          any `path:"/settings"`
	SettingCat string
	SettingId  string
}

type Query struct {
	Foo *string
	Bar int
}

type Body struct {
	Value any     `json:"value"`
	Type  string  `json:"type"`
	Opt   *string `json:"opt"`
}

type Headers struct {
	Yo          any
	ContentType string `name:"Content-Type"`
}

type RespHeaders struct {
	ContentType string `name:"Content-Type"`
}

var Get = Endpoint[
	EndpointInput[Headers, Path, Query, X],
	EndpointOutput[X, Body],
]{Method: http.MethodGet}

var Put = Endpoint[
	EndpointInput[X, Path, X, Body],
	EndpointOutput[RespHeaders, X],
]{Method: http.MethodPost}
