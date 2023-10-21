package apio

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////
///// PUBLIC FOUNDATION TYPES

type Api struct {
	Name        string
	Servers     []Server
	IntBasePath string
	Endpoints   []EndpointBase
}

type Server struct {
	Scheme   string
	Host     string
	Port     int
	BasePath string
	HttpVer  string
}

func (a Api) WithEndpoints(endpoint ...EndpointBase) Api {
	a.Endpoints = append(a.Endpoints, endpoint...)
	return a
}

func (a Api) Validate() Api {
	for _, e := range a.Endpoints {
		e.validate()
	}
	return a
}

type ErrResp struct {
	Status int
	ClMsg  string
	IntErr error
}

func (e ErrResp) Error() string {
	return fmt.Sprintf("err response: %d: %s", e.Status, e.ClMsg)
}

func NewError(code int, message string, intErr error) error {
	return &ErrResp{
		Status: code,
		ClMsg:  message,
		IntErr: intErr,
	}
}

type EndpointBase interface {
	getMethod() string
	getPathPattern() string
	getQueryPattern() string
	invoke(payload inputPayload) (endpointOutputBase, error)
	validate()
}

type Endpoint[Input endpointInputBase, Output endpointOutputBase] struct {
	Method         string
	Handler        func(Input) (Output, error)
	headerBindings *HeaderBindings
	pathBindings   *PathBindings
	queryBindings  *QueryBindings
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

type EndpointOutput[
	HeadersType any,
	BodyType any,
] struct {
	Code    int
	Headers HeadersType
	Body    BodyType
}

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////
///// PUBLIC HELPER TYPES (responses)

type X struct{}

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

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////
///// PRIVATE IMPL

func (e Endpoint[Input, Output]) getHeaderBindings() HeaderBindings {
	if e.headerBindings == nil {
		b := calcHeaderBindings[Input]()
		e.headerBindings = &b
	}
	return *e.headerBindings
}

func (e Endpoint[Input, Output]) getPathBindings() PathBindings {
	if e.pathBindings == nil {
		b := calcPathBindings[Input]()
		e.pathBindings = &b
	}
	return *e.pathBindings
}

func (e Endpoint[Input, Output]) getQueryBindings() QueryBindings {
	if e.queryBindings == nil {
		b := calcQueryBindings[Input]()
		e.queryBindings = &b
	}
	return *e.queryBindings
}

func (e Endpoint[Input, Output]) validateInputBodyType() {
	// We only check that it is a struct
	var zero Input
	zero.validateBodyType()
}

func (e Endpoint[Input, Output]) validateOutputBodyType() {
	var zero Output
	zero.validateBodyType()
}

func (e Endpoint[Input, Output]) validateOutputHeadersType() {
	var zero Output
	zero.validateHeadersType()
}

func (e Endpoint[Input, Output]) invoke(payload inputPayload) (endpointOutputBase, error) {
	var zeroInput Input
	var zeroOutput Output
	input, err := zeroInput.parse(payload, e.getHeaderBindings(), e.getPathBindings(), e.getQueryBindings())
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
		} else {
			return zeroOutput, NewError(http.StatusInternalServerError, fmt.Sprintf("failed to invoke handler: %v", err), err)
		}
	}
	return output, nil
}

func AsErResp(err error) *ErrResp {
	var errResp *ErrResp
	if errors.As(err, &errResp) {
		return errResp
	}
	return nil
}

func (e Endpoint[Input, Output]) getMethod() string {
	return e.Method
}

func (e Endpoint[Input, Output]) getPathPattern() string {
	return e.getPathBindings().FlatPath
}

func (e Endpoint[Input, Output]) getQueryPattern() string {
	return e.getQueryBindings().FlatPath
}

func (e Endpoint[Input, Output]) validate() {
	e.getHeaderBindings() // panics if invalid
	e.getPathBindings()   // panics if invalid
	e.getQueryBindings()  // panics if invalid
	e.validateInputBodyType()
	e.validateOutputBodyType()
	e.validateOutputHeadersType()
}
