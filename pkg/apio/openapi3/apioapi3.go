package openapi3

import (
	"github.com/GiGurra/apio/pkg/apio"
	"strconv"
)

type OpenApi struct {
	Openapi    string            `json:"openapi" yaml:"openapi" text:"openapi"`
	Info       map[string]string `json:"info" yaml:"info" text:"info"`
	Servers    []Server          `json:"servers" yaml:"servers" text:"servers"`
	Paths      map[string]any    `json:"paths" yaml:"paths" text:"paths"`
	Components map[string]any    `json:"components" yaml:"components" text:"components"`
}

type Server struct {
	URL         string                 `json:"url" yaml:"url" text:"url"`
	Description string                 `json:"description" yaml:"description" text:"description"`
	Variables   map[string]interface{} `json:"variables,omitempty" yaml:"variables" text:"variables"`
}

type Schema struct {
	Type       string         `json:"type" yaml:"type" text:"type"`
	Properties map[string]any `json:"properties" yaml:"properties" text:"properties"`
}
type Parameter struct {
	Name        string         `json:"name" yaml:"name" text:"name"`
	In          string         `json:"in" yaml:"in" text:"in"`
	Description string         `json:"description" yaml:"description" text:"description"`
	Required    bool           `json:"required" yaml:"required" text:"required"`
	Schema      map[string]any `json:"schema" yaml:"schema" text:"schema"`
}

type Response struct {
	Description string         `json:"description" yaml:"description" text:"description"`
	Content     map[string]any `json:"content" yaml:"content" text:"content"`
}

type Operation struct {
	Summary     string              `json:"summary" yaml:"summary" text:"summary"`
	Description string              `json:"description" yaml:"description" text:"description"`
	OperationId string              `json:"operationId" yaml:"operationId" text:"operation_id"`
	Parameters  []Parameter         `json:"parameters" yaml:"parameters" text:"parameters"`
	Responses   map[string]Response `json:"responses" yaml:"responses" text:"responses"`
}

func ToOpenApi3(api apio.Api) OpenApi {

	servers := make([]Server, len(api.Servers))
	for i, server := range api.Servers {
		servers[i] = Server{
			URL:         server.Scheme + "://" + server.Host + ":" + strconv.Itoa(server.Port) + server.BasePath,
			Description: server.Name + " - " + server.Description,
			Variables:   server.Properties(),
		}
	}

	return OpenApi{
		Openapi: "3.0.0",
		Info: map[string]string{
			"title":       api.Name,
			"description": api.Description,
			"version":     api.Version,
		},
		Servers:    servers,
		Paths:      GetPaths(api),
		Components: GetComponents(api),
	}
}

func GetPaths(api apio.Api) map[string]any {
	return map[string]any{
		"/users": map[string]Operation{
			"get": {
				Summary:     "List all users",
				Description: "This operation lists all users in the system",
				OperationId: "listUsers",
				Parameters:  []Parameter{},
				Responses: map[string]Response{
					"200": {
						Description: "A list of users",
						Content: map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "array",
									"items": map[string]any{
										"$ref": "#/components/schemas/User",
									},
								},
							},
						},
					},
				},
			},
		},
		"/users/{id}": map[string]Operation{
			"get": {
				Summary:     "Get a user by ID",
				Description: "This operation retrieves a user by ID",
				OperationId: "getUser",
				Parameters: []Parameter{
					{
						Name:        "id",
						In:          "path",
						Description: "User ID",
						Required:    true,
						Schema: map[string]any{
							"type": "integer",
						},
					},
				},
				Responses: map[string]Response{
					"200": {
						Description: "A user",
						Content: map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"$ref": "#/components/schemas/User",
								},
							},
						},
					},
				},
			},
		},
	}
}

func GetComponents(api apio.Api) map[string]any {
	return map[string]any{
		"schemas": map[string]any{
			"User": Schema{
				Type: "object",
				Properties: map[string]any{
					"id": map[string]any{
						"type": "integer",
					},
					"name": map[string]any{
						"type": "string",
					},
					"email": map[string]any{
						"type": "string",
					},
				},
			},
		},
	}
}
