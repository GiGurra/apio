package openapi3

import (
	"encoding/json"
	"fmt"
	"github.com/GiGurra/apio/pkg/apio"
	"github.com/google/go-cmp/cmp"
	"net/http"
	"testing"
)

func TestBodyIsSlice(t *testing.T) {

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
		apio.EndpointOutput[X, []User],
	]{
		Method:      http.MethodGet,
		ID:          "GetUser",
		Name:        "GetUser",
		Summary:     "Get a user by id",
		Description: "Get a user by id, the long description.",
		Tags:        []string{"Users"},
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

	refJson := make(map[string]any)
	err = json.Unmarshal([]byte(refSliceJsonStr), &refJson)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to unmarshal reference OpenAPI 3 spec: %v", err))
	}

	actualJson := make(map[string]any)
	err = json.Unmarshal(openApi3Str, &actualJson)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to unmarshal actual OpenAPI 3 spec: %v", err))
	}

	if diff := cmp.Diff(refJson, actualJson); diff != "" {
		t.Fatalf("OpenAPI 3 spec mismatch:\n%s", diff)
	}

	fmt.Printf("actualJson: %s\n", actualJson)
}

var refSliceJsonStr = `
{
  "openapi": "3.0.0",
  "info": {
    "description": "This is a test API.",
    "title": "My test API",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "https://api.example.com:443/api/v1",
      "description": "My test server - My test server description"
    }
  ],
  "paths": {
    "/users/{User}": {
      "get": {
        "summary": "Get a user by id",
        "description": "Get a user by id, the long description.",
        "operationId": "GetUser",
        "parameters": [
          {
            "name": "Yo",
            "in": "header",
            "description": "Yo",
            "required": true,
            "schema": {
              "type": "string"
            }
          },
          {
            "name": "User",
            "in": "path",
            "description": "User",
            "required": true,
            "schema": {
              "type": "integer"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "",
            "content": {
              "application/json": {
                "schema": {
                  "items": {
                    "$ref": "#/components/schemas/openapi3_User"
                  },
                  "type": "array"
                }
              }
            }
          }
        },
        "tags": [
          "Users"
        ]
      }
    }
  },
  "components": {
    "schemas": {
      "openapi3_Address": {
        "type": "object",
        "properties": {
          "City": {
            "type": "string"
          },
          "Street": {
            "type": "string"
          },
          "ZipCode": {
            "type": "string"
          }
        },
        "required": [
          "Street",
          "City",
          "ZipCode"
        ]
      },
      "openapi3_User": {
        "type": "object",
        "properties": {
          "Address": {
            "$ref": "#/components/schemas/openapi3_Address"
          },
          "Email": {
            "type": "string"
          },
          "ExtraAddresses": {
            "items": {
              "$ref": "#/components/schemas/openapi3_Address"
            },
            "type": "array"
          },
          "Name": {
            "type": "string"
          }
        },
        "required": [
          "Name",
          "Email",
          "Address",
          "ExtraAddresses"
        ]
      }
    }
  }
}`
