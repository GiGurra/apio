package apio

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type Payload struct {
	Headers map[string][]string
	Path    map[string]string
	Query   map[string][]string
	Body    []byte
}

type EndpointBase interface {
	GetMethod() string
	GetPath() string
	Invoke(payload Payload) (EndpointOutputBase, error)
}

type EndpointInputBase interface {
	GetHeaders() any
	GetPath() any
	CalcPathBinding() PathBinding
	GetQuery() any
	GetBody() any
	Parse(payload Payload, pathBinding PathBinding) (any, error)
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

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) Parse(
	payload Payload,
	pathBinding PathBinding,
) (any, error) {
	fmt.Printf("todo: implement EndpointInput.Parse\n")
	var result EndpointInput[HeadersType, PathType, QueryType, BodyType]
	for name, setter := range pathBinding.Bindings {
		inputValue, ok := payload.Path[name]
		if !ok {
			return result, fmt.Errorf("missing path parameter '%s'", name)
		}
		valueToSet := reflect.ValueOf(result.Path).Elem().FieldByName(name)
		err := setter(valueToSet, inputValue)
		if err != nil {
			return result, fmt.Errorf("failed to set path parameter '%s': %w", name, err)
		}
	}
	return result, nil
}

type PathBinding struct {
	FlatPath string
	Bindings map[string]fieldSetter
}

func (p PathBinding) BindData(instance any, params map[string]string) error {
	elem := reflect.ValueOf(instance).Elem()
	for key, value := range params {
		if setter, ok := p.Bindings[key]; ok {
			field := elem.FieldByName(key)
			err := setter(field, value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type fieldSetter = func(v reflect.Value, value string) error

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) CalcPathBinding() PathBinding {

	// iterate over fields in PathType

	// Check that PathType is of type struct
	pathT := reflect.TypeOf((*PathType)(nil)).Elem()

	if pathT.Kind() != reflect.Struct {
		panic("PathType must be a struct")
	}

	result := PathBinding{
		FlatPath: "",
		Bindings: make(map[string]fieldSetter),
	}
	alreadyTaken := make(map[string]bool)

	// Iterate over fields in PathType
	for i := 0; i < pathT.NumField(); i++ {
		// Check if the field has path
		field := pathT.Field(i)

		setter := func() fieldSetter {
			switch field.Type.Kind() {
			case reflect.String:
				return func(v reflect.Value, value string) error {
					v.SetString(value)
					return nil
				}
			default:
				// use the json mapping
				return func(v reflect.Value, value string) error {
					err := json.Unmarshal([]byte(value), v.Interface())
					if err != nil {
						return fmt.Errorf("failed to unmarshal '%s' into '%s': %w", value, field.Type.Name(), err)
					}
					return nil
				}
			}
		}()

		if field.Name == "_" {
			// We won't bind this parameter, but it is still needed in the path
			// Check if it has a tag called path
			pathTag := field.Tag.Get("path")
			if pathTag == "" {
				// Treat as wildcard
				result.FlatPath += "/*"
			} else {
				// Treat as literal
				result.FlatPath += "/" + strings.TrimPrefix(pathTag, "/")
			}
		} else {
			if alreadyTaken[field.Name] {
				panic(fmt.Sprintf("field '%s' is already taken", field.Name))
			}

			alreadyTaken[field.Name] = true
			result.FlatPath += "/:" + field.Name
			result.Bindings[field.Name] = setter
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

type EndpointOutputBase interface {
	GetCode() int
	GetHeaders() map[string][]string
	GetBody() ([]byte, error)
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

func (e EndpointOutput[HeadersType, BodyType]) GetHeaders() map[string][]string {
	fmt.Printf("TODO: implement EndpointOutput.GetHeaders\n")
	return make(map[string][]string)
}

func (e EndpointOutput[HeadersType, BodyType]) GetBody() ([]byte, error) {
	bytes, err := json.Marshal(e.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}
	return bytes, nil
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

type Endpoint[Input EndpointInputBase, Output EndpointOutputBase] struct {
	Method      string
	Handler     func(Input) (Output, error)
	pathBinding *PathBinding
}

func calcPathBinding[Input EndpointInputBase]() PathBinding {
	var zero Input
	return zero.CalcPathBinding()
}

func (e Endpoint[Input, Output]) WithMethod(method string) Endpoint[Input, Output] {
	e.Method = method
	return e
}

func (e Endpoint[Input, Output]) GetPathBinding() PathBinding {
	if e.pathBinding == nil {
		b := calcPathBinding[Input]()
		e.pathBinding = &b
	}
	return *e.pathBinding
}

func (e Endpoint[Input, Output]) Invoke(payload Payload) (EndpointOutputBase, error) {
	var zeroInput Input
	var zeroOutput Output
	input, err := zeroInput.Parse(payload, e.GetPathBinding())
	if err != nil {
		return zeroOutput, NewError(http.StatusBadRequest, fmt.Sprintf("failed to parse input: %v", err), err)
	}
	inputAsInput, ok := input.(Input)
	if !ok {
		return zeroOutput, NewError(http.StatusInternalServerError, fmt.Sprintf("failed to cast input to %t", reflect.TypeOf(zeroInput)), nil)
	}
	output, err := e.Handler(inputAsInput)
	if err != nil {
		var errResp *ErrResp
		if errors.As(err, &errResp) {
			return zeroOutput, errResp
		}
	}
	return output, nil
}

func (e Endpoint[Input, Output]) WithHandler(handler func(Input) (Output, error)) Endpoint[Input, Output] {
	e.Handler = handler
	return e
}

func (e Endpoint[Input, Output]) GetMethod() string {
	return e.Method
}

func (e Endpoint[Input, Output]) GetPath() string {
	return e.GetPathBinding().FlatPath
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
