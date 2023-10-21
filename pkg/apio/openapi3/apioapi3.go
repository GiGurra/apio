package openapi3

import (
	"github.com/GiGurra/apio/pkg/apio"
	"reflect"
	"strconv"
	"strings"
)

type OpenApi struct {
	Openapi    string            `json:"openapi" yaml:"openapi" text:"openapi"`
	Info       map[string]string `json:"info" yaml:"info" text:"info"`
	Servers    []Server          `json:"servers" yaml:"servers" text:"servers"`
	Paths      map[string]any    `json:"paths" yaml:"paths" text:"paths"`
	Components map[string]any    `json:"components" yaml:"components" text:"components"`
}

type Server struct {
	URL         string `json:"url" yaml:"url" text:"url"`
	Description string `json:"description" yaml:"description" text:"description"`
}

type Schema struct {
	Type       string         `json:"type" yaml:"type" text:"type"`
	Properties map[string]any `json:"properties" yaml:"properties" text:"properties"`
	Required   []string       `json:"required" yaml:"required" text:"required"`
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

func apioPattern2OpenApi3Pattern(pattern string) string {

	result := ""
	for _, part := range strings.Split(pattern, "/") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if trimmed[0] == ':' {
			result += "/{" + trimmed[1:] + "}"
		} else {
			result += "/" + trimmed
		}
	}
	return result
}

func goTypeToOpenapiType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int:
		return "integer"
	case reflect.Int8:
		return "integer"
	case reflect.Int16:
		return "integer"
	case reflect.Int32:
		return "integer"
	case reflect.Int64:
		return "integer"
	case reflect.Uint:
		return "integer"
	case reflect.Uint8:
		return "integer"
	case reflect.Uint16:
		return "integer"
	case reflect.Uint32:
		return "integer"
	case reflect.Uint64:
		return "integer"
	case reflect.Float32:
		return "number"
	case reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Interface:
		fallthrough
	case reflect.Struct:
		return "object"
	case reflect.Ptr:
		return goTypeToOpenapiType(t.Elem())
	case reflect.Slice:
		return "array"
	default:
		panic("unsupported type: " + t.String())
	}
}

func GetParameters(api apio.EndpointBase) []Parameter {
	result := make([]Parameter, 0)

	for _, field := range api.GetInputHeaderInfo().Fields {

		if field.LKName == "content-type" {
			continue // OpenAPI 3 spec doesn't permit this here
		}

		result = append(result, Parameter{
			Name:        field.Name,
			In:          "header",
			Description: field.Name,
			Required:    true,
			Schema: map[string]any{
				"type": goTypeToOpenapiType(field.ValueType),
			},
		})
	}

	for _, field := range api.GetInputPathInfo().Fields {
		result = append(result, Parameter{
			Name:        field.Name,
			In:          "path",
			Description: field.Name,
			Required:    true,
			Schema: map[string]any{
				"type": goTypeToOpenapiType(field.ValueType),
			},
		})
	}

	for _, field := range api.GetInputQueryInfo().Fields {
		result = append(result, Parameter{
			Name:        field.Name,
			In:          "query",
			Description: field.Name,
			Required:    field.IsRequired(),
			Schema: map[string]any{
				"type": goTypeToOpenapiType(field.ValueType),
			},
		})
	}

	return result
}

func GetPaths(api apio.Api) map[string]any {

	result := make(map[string]any)
	for _, e := range api.Endpoints {
		path := apioPattern2OpenApi3Pattern(e.GetPathPattern())
		if path[0] != '/' {
			path = "/" + path
		}
		if _, ok := result[path]; !ok {
			result[path] = make(map[string]any)
		}
		methods := result[path].(map[string]any)

		bodyInfo := e.GetBodyOutputInfo()
		content := make(map[string]any)
		if bodyInfo.HasContent() {
			content = map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"$ref": "#/components/schemas/" + schemaNameOf(bodyInfo),
					},
				},
			}
		}

		methods[strings.ToLower(e.GetMethod())] = Operation{
			Summary:     e.GetSummary(),
			Description: e.GetDescription(),
			OperationId: e.GetId(),
			Parameters:  GetParameters(e),
			Responses: map[string]Response{
				strconv.Itoa(e.OkCode()): {
					Description: e.GetOutput().GetDescription(),
					Content:     content,
				},
			},
		}
	}

	return result
}

func GetComponents(api apio.Api) map[string]any {

	schemas := make(map[string]any)
	for _, e := range api.Endpoints {
		bodyInfos := []apio.AnalyzedStruct{
			e.GetBodyOutputInfo(),
			e.GetBodyInputInfo(),
		}
		for _, bodyInfo := range bodyInfos {
			if bodyInfo.HasContent() {
				props := make(map[string]any)
				required := make([]string, 0)

				for _, field := range bodyInfo.Fields {
					props[field.Name] = map[string]any{
						"type": goTypeToOpenapiType(field.ValueType),
					}
					if field.IsRequired() {
						required = append(required, field.Name)
					}
				}

				schemas[schemaNameOf(bodyInfo)] = Schema{
					Type:       "object",
					Properties: props,
					Required:   required,
				}
			}
		}
	}

	return map[string]any{
		"schemas": schemas,
	}
}

func schemaNameOf(structInfo apio.AnalyzedStruct) string {
	pkgParts := strings.Split(structInfo.Pkg, "/")
	lastPart := pkgParts[len(pkgParts)-1]
	raw := lastPart + "/" + structInfo.Name
	// keep only alphanumeric characters, underscores and '-'s. Replace all other characters with '_'
	return strings.Map(func(r rune) rune {
		if r == '-' || r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= 9) {
			return r
		}
		return '_'
	}, raw)
}
