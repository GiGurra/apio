package openapi3

import (
	"encoding/json"
	"fmt"
	"github.com/GiGurra/apio/pkg/apio"
	"net/http"
	"testing"
)

func TestToOpenApi3(t *testing.T) {

	type InputHeaders struct {
		Yo          string
		ContentType string `name:"Content-Type"`
	}

	type InputPath struct {
		_    any `path:"/users"`
		User int
	}
	type Address struct {
		Street  string
		City    string
		ZipCode string
	}

	type User struct {
		Name           string
		Email          string
		Address        Address
		ExtraAddresses []Address
	}

	type X = apio.X

	endpoint := apio.Endpoint[
		apio.EndpointInput[InputHeaders, InputPath, X, X],
		apio.EndpointOutput[X, User],
	]{
		Method:      http.MethodGet,
		ID:          "GetUser",
		Name:        "GetUser",
		Summary:     "GetUser",
		Description: "GetUser",
	}

	testApi := apio.Api{
		Name:        "My test API",
		Description: "This is a test API.",
		Version:     "1.0.0",

		Servers: []apio.Server{{
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
		endpoint,
	).Validate()

	openApi3Str, err := json.MarshalIndent(ToOpenApi3(testApi), "", "  ")
	if err != nil {
		t.Fatal(fmt.Errorf("failed to marshal OpenAPI 3 spec: %v", err))
	}
	fmt.Printf("openApi3: %s\n", openApi3Str)
}
