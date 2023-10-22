package user_setting

import (
	. "github.com/GiGurra/apio/pkg/apio"
	"net/http"
)

type PathAll struct {
	_    any `path:"/users"`
	User int
	_    any `path:"/settings"`
}

// PathById represents "/users/:user/settings/:settingCat/:settingId"
type PathById struct {
	_          any `path:"/users"`
	User       int
	_          any `path:"/settings"`
	SettingCat string
	SettingId  string
}

type PathByCat struct {
	_          any `path:"/users"`
	User       int
	_          any `path:"/settings"`
	SettingCat string
}

type Query struct {
	Foo *string
	Bar int
}

type Body struct {
	Value string  `json:"value"`
	Type  string  `json:"type"`
	Opt   *string `json:"opt"`
}

type Headers struct {
	Yo          string
	ContentType string `name:"Content-Type"`
}

type RespHeaders struct {
	ContentType string `name:"Content-Type"`
}

var GetAll = Endpoint[
	EndpointInput[Headers, PathAll, X, X],
	EndpointOutput[RespHeaders, []Body],
]{
	Method:      http.MethodGet,
	ID:          "getUserSettings",
	Summary:     "Get all user setting",
	Description: "This operation retrieves all user settings",
	Tags:        []string{"Users"},
}

var GetById = Endpoint[
	EndpointInput[Headers, PathById, Query, X],
	EndpointOutput[RespHeaders, Body],
]{
	Method:      http.MethodGet,
	ID:          "getUserSetting",
	Summary:     "Get a user setting",
	Description: "This operation retrieves a user setting",
	Tags:        []string{"Users"},
}

var PutById = Endpoint[
	EndpointInput[X, PathById, X, Body],
	EndpointOutput[RespHeaders, X],
]{
	Method:      http.MethodPut,
	ID:          "putUserSetting",
	Summary:     "Replace a user setting",
	Description: "This operation replaces a user setting",
	Tags:        []string{"Users"},
}
