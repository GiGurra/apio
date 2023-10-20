package apio

type Payload interface {
}

type EndpointBase interface {
	GetMethod() string
	GetPath() string
}

type EndpointInput[
	HeaderType any,
	PathType any,
	QueryType any,
	BodyType any,
] struct {
}

type EndpointOutput[
	HeaderType any,
	PathType any,
	QueryType any,
	BodyType any,
] struct {
}

type Endpoint[Input, Output any] struct {
	Method string
	Path   string
}

func (e *Endpoint[Input, Output]) GetMethod() string {
	return e.Method
}

func (e *Endpoint[Input, Output]) GetPath() string {
	return e.Path
}

type Server struct {
	Scheme   string
	Host     string
	Port     int
	BasePath string
}

type Api struct {
	Servers   []Server
	Endpoints []EndpointBase
}
