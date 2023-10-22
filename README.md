# apio

This is a **very** experimental declarative server- and client-side http API library for Go.
It is inspired by Tapir (scala). Again, this is an experiment. You should **_probably_** just generate your server and
client from an OpenAPI spec instead.

This library asks the question:

* What do we need to describe an API using the Go type system?

It lets you:

* Define your API with declarative Go code & types
* Get direct access to type safe client and server (and in tests).
* Validate your API definition at application startup (or in tests).
* Optionally generate an OpenAPI/swagger spec from the API definition.

Without any required Go code generation.

## Example

### API definition

An API consists of one or more resources, each with one or more endpoints.
The simplest way of grouping the endpoints together is to simply have a go
package for each resource:

```go
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
	EndpointInput[X, PathAll, Query, X],
	EndpointOutput[RespHeaders, []Body],
]{
	Method:      http.MethodGet,
	ID:          "getUserSettings",
	Summary:     "Get all user setting",
	Description: "This operation retrieves all user settings",
	Tags:        []string{"Users"},
}

var GetById = Endpoint[
	EndpointInput[X, PathById, X, X],
	EndpointOutput[RespHeaders, Body],
]{
	Method:      http.MethodGet,
	ID:          "getUserSetting",
	Summary:     "Get a user setting",
	Description: "This operation retrieves a user setting",
	Tags:        []string{"Users"},
}

var PutById = Endpoint[
	EndpointInput[Headers, PathById, X, Body],
	EndpointOutput[X, X],
]{
	Method:      http.MethodPut,
	ID:          "putUserSetting",
	Summary:     "Replace a user setting",
	Description: "This operation replaces a user setting",
	Tags:        []string{"Users"},
}

```

### Server

Then we create a server. We use the previously defined endpoints specifications and attach handlers to each.
We can do this in a single source file or split/grouped into multiple files, whatever you prefer.

Here is a simple example using an Echo Server:
(it could be any server implementation)

```go
package main

import (
	"fmt"
	"github.com/GiGurra/apio/cmd/examples/user_setting"
	. "github.com/GiGurra/apio/pkg/apio"
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

	echoServer := echo.New()

	EchoInstall(echoServer, &testApi)

	_ = echoServer.Start(":8080")
}

```

### Client

Similar to how we created the server, we can use the api endpoint specifications to make requests.
`apio` has a built-in implementation using the go standard library http client, but you could also use your own.
All the data required is public. See the default implementation for reference.

Here is an example using the built-in functionality:

```go
package main

import (
	"fmt"
	"github.com/GiGurra/apio/cmd/examples/user_setting"
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

	input1 := NewInput(
		Empty, // no headers in this get call
		user_setting.PathById{
			User:       123,
			SettingCat: "cat",
			SettingId:  "id",
		},
		Empty, // no query in this get call
		Empty, // no body in this get call
	)

	res1, err1 := user_setting.GetById.RPC(server, input1, DefaultOpts())
	if err1 != nil {
		panic(fmt.Sprintf("failed to call RPC GET endpoint: %v", err1))
	}

	fmt.Printf("res: %+v\n", res1)

	input2 := NewInput(
		Empty, // no headers in this get call
		user_setting.PathAll{
			User: 123,
		},
		user_setting.Query{
			Foo: ptr("foo"),
			Bar: 123,
		},
		Empty, // no body in this get call
	)

	res2, err2 := user_setting.GetAll.RPC(server, input2, DefaultOpts())
	if err2 != nil {
		panic(fmt.Sprintf("failed to call RPC GET all endpoint: %v", err1))
	}

	fmt.Printf("res: %+v\n", res2)
}

```

### OpenAPI 3 spec

We can also generate an OpenAPI 3 spec from the API definition:

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/GiGurra/apio/cmd/examples/user_setting"
	. "github.com/GiGurra/apio/pkg/apio"
	"github.com/GiGurra/apio/pkg/apio/openapi3"
	"github.com/labstack/echo/v4"
)

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
	).Validate()

	openApi3 := openapi3.ToOpenApi3(testApi)
	openApi3Json, err := json.MarshalIndent(openApi3, "", "  ")
	if err != nil {
		panic(fmt.Errorf("failed to marshal OpenAPI 3 spec: %v", err))
	}

	fmt.Printf("OpenAPI 3 spec:\n")
	fmt.Printf("%s\n", openApi3Json) // Paste this in your favorite OpenAPI 3 editor

}


```

![img.png](img.png)

## TODO

* [ ] Add support for struct composition
* [ ] Add support for other content types than JSON

