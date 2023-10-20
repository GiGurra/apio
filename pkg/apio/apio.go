package apio

type Payload interface {
}

type EndpointBase interface {
	GetMethod() string
	GetPath() string
}

type EndpointInput[
	Headers any,
	Path any,
	Query any,
	Body any,
] struct {
}

type Empty struct{}

type NoHeaders struct{}

type NoPath struct{}

type NoQuery struct{}

type NoBody struct{}

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
