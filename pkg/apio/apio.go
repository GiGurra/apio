package apio

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

type EndpointBase interface {
	getMethod() string
	getPath() string
	invoke(payload inputPayload) (EndpointOutputBase, error)
}

type inputPayload struct {
	Headers map[string][]string
	Path    map[string]string
	Query   map[string][]string
	Body    []byte
}

type X any

func EmptyResponse() EndpointOutput[X, X] {
	return EndpointOutput[X, X]{
		Code: 204,
	}
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

type Endpoint[Input endpointInputBase, Output EndpointOutputBase] struct {
	Method      string
	Handler     func(Input) (Output, error)
	pathBinding *PathBinding
}

func calcPathBinding[Input endpointInputBase]() PathBinding {
	var zero Input
	return zero.calcPathBinding()
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

func (e Endpoint[Input, Output]) invoke(payload inputPayload) (EndpointOutputBase, error) {
	var zeroInput Input
	var zeroOutput Output
	input, err := zeroInput.parse(payload, e.GetPathBinding())
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

func (e Endpoint[Input, Output]) getMethod() string {
	return e.Method
}

func (e Endpoint[Input, Output]) getPath() string {
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
