# apio

This is a **very** experimental declarative server- and client-side http API library for Go.
It is inspired by Tapir (scala).

You define your API in a type-safe way with declarative Go code, and then
get access to type safe client and server APIs in go (no code generation).

You can also (in theory ;)) generate a swagger spec from the API definition.

It asks the question:

* What do we need to describe an API entirely within the Go type system, but to be able to always generate a valid API
  specification from it? (ofc it also has to work :D)

## Example

First we specify one or more endpoints

```go
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
```

Then we start a server:

```go
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

```

Then we can use the client:

```go
package main

import (
	"fmt"
	. "github.com/GiGurra/apio/cmd/examples/common"
	. "github.com/GiGurra/apio/pkg/apio"
)

func ptr[T any](v T) *T {
	return &v
}

func main() {
	server := Server{
		Scheme:   "http",
		Host:     "localhost",
		Port:     8080,
		BasePath: "/api/v1",
		HttpVer:  "1.1",
	}

	input := NewInput(
		UserSettingHeaders{
			Yo:          "yo",
			ContentType: "application/json",
		},
		UserSettingPath{
			User:       123,
			SettingCat: "cat",
			SettingId:  "id",
		},
		UserSettingQuery{
			Foo: ptr("foo"),
			Bar: 123,
		},
		Empty, // No body sent (this is a GET call)
	)

	res, err := GetEndpointSpec.RPC(server, input, DefaultOpts())
	if err != nil {
		panic(fmt.Sprintf("failed to call RPC GET endpoint: %v", err))
	}

	fmt.Printf("res: %+v\n", res)
}

```