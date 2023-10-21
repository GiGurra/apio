package apio

import (
	"fmt"
	"reflect"
	"strings"
)

type Payload interface {
}

type EndpointBase interface {
	GetMethod() string
	GetPath() string
}

type EndpointInputBase interface {
	GetHeaders() any
	GetPath() any
	GetPathPattern() string
	GetQuery() any
	GetBody() any
}

type EndpointInput[
	HeadersType any,
	PathType any,
	QueryType any,
	BodyType any,
] struct {
	Headers HeadersType
	Path    PathType
	Query   QueryType
	Body    BodyType
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) GetHeaders() any {
	return e.Headers
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) GetPath() any {
	return e.Path
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) GetPathPattern() string {

	// iterate over fields in PathType

	// Check that PathType is of type struct
	pathT := reflect.TypeOf((*PathType)(nil)).Elem()

	if pathT.Kind() != reflect.Struct {
		panic("PathType must be a struct")
	}

	result := ""
	alreadyTaken := make(map[string]bool)

	// Iterate over fields in PathType
	for i := 0; i < pathT.NumField(); i++ {
		// Check if the field has path
		field := pathT.Field(i)
		if field.Name == "_" {
			// We won't bind this parameter, but it is still needed in the path
			// Check if it has a tag called path
			pathTag := field.Tag.Get("path")
			if pathTag == "" {
				// Treat as wildcard
				result += "/*"
			} else {
				// Treat as literal
				result += "/" + strings.TrimPrefix(pathTag, "/")
			}
		} else {
			if alreadyTaken[field.Name] {
				panic(fmt.Sprintf("field '%s' is already taken", field.Name))
			}

			alreadyTaken[field.Name] = true
			result += "/:" + field.Name
		}
	}

	return result
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) GetQuery() any {
	return e.Query
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) GetBody() any {
	return e.Body
}

type X any

type EndpoitOutputBase interface {
	GetCode() int
	GetHeaders() any
	GetBody() any
}

type EndpointOutput[
	HeadersType any,
	BodyType any,
] struct {
	Code    int
	Headers HeadersType
	Body    BodyType
}

func (e EndpointOutput[HeadersType, BodyType]) GetCode() int {
	return e.Code
}

func (e EndpointOutput[HeadersType, BodyType]) GetHeaders() any {
	return e.Headers
}

func (e EndpointOutput[HeadersType, BodyType]) GetBody() any {
	return e.Body
}

func EmptyResponse() EndpointOutput[X, X] {
	return EndpointOutput[X, X]{}
}

func Response[H any, B any](headers H, body B) EndpointOutput[H, B] {
	return EndpointOutput[H, B]{
		Code:    200,
		Headers: headers,
		Body:    body,
	}
}

func BodyResponse[BodyType any](body BodyType) EndpointOutput[X, BodyType] {
	return EndpointOutput[X, BodyType]{
		Code: 200,
		Body: body,
	}
}

func HeadersResponse[H any](headers H) EndpointOutput[H, X] {
	return EndpointOutput[H, X]{
		Code:    204,
		Headers: headers,
	}
}

type Endpoint[Input EndpointInputBase, Output EndpoitOutputBase] struct {
	Method     string
	Handler    func(Input) (Output, error)
	cachedPath string
}

func calcPathPattern[Input EndpointInputBase]() string {
	var zero Input
	return zero.GetPathPattern()
}

func (e Endpoint[Input, Output]) WithMethod(method string) Endpoint[Input, Output] {
	e.Method = method
	return e
}

func (e Endpoint[Input, Output]) WithHandler(handler func(Input) (Output, error)) Endpoint[Input, Output] {
	e.Handler = handler
	return e
}

func (e Endpoint[Input, Output]) GetMethod() string {
	return e.Method
}

func (e Endpoint[Input, Output]) GetPath() string {
	if e.cachedPath == "" {
		e.cachedPath = calcPathPattern[Input]()
	}
	return e.cachedPath
}

type Server struct {
	Scheme   string
	Host     string
	Port     int
	BasePath string
	HttpVer  string
}

type Api struct {
	Published   []Server
	IntBasePath string
	Endpoints   []EndpointBase
}

func (a Api) AddEndpoints(endpoint ...EndpointBase) Api {
	a.Endpoints = append(a.Endpoints, endpoint...)
	return a
}

type ErrResp struct {
	Code    int
	Message string
	IntErr  error
}

func (e ErrResp) Error() string {
	return fmt.Sprintf("err response: %d: %s", e.Code, e.Message)
}

func NewError(code int, message string, intErr error) error {
	return &ErrResp{
		Code:    code,
		Message: message,
		IntErr:  intErr,
	}
}
