package apio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type InputPayload struct {
	Headers map[string][]string
	Path    map[string]string
	PathStr string
	Query   map[string][]string
	Body    []byte
}

func (p InputPayload) QueryString() string {
	if len(p.Query) == 0 {
		return ""
	}
	var result strings.Builder
	result.WriteString("?")
	for k, v := range p.Query {
		urlEncodedValue := url.QueryEscape(v[0])
		result.WriteString(k + "=" + urlEncodedValue + "&")
	}
	return strings.TrimSuffix(result.String(), "&")
}

type EndpointInputBase interface {
	getHeaders() any
	getPath() any
	calcHeaderBindings() HeaderBindings
	calcPathBindings() PathBindings
	calcQueryBindings() QueryBindings
	validateBodyType()
	getQuery() any
	getBody() any
	parse(
		payload InputPayload,
		bindings HeaderBindings,
		pathBinding PathBindings,
		queryBindings QueryBindings,
	) (any, error)
	ToPayload() (InputPayload, error)
	GetHeaderInfo() AnalyzedStruct
	GetPathInfo() AnalyzedStruct
	GetQueryInfo() AnalyzedStruct
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) GetHeaderInfo() AnalyzedStruct {
	info, err := AnalyzeStruct(e.Headers)
	if err != nil {
		panic(fmt.Errorf("failed to analyze headers: %w", err))
	}
	return info
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) GetPathInfo() AnalyzedStruct {
	info, err := AnalyzeStruct(e.Path)
	if err != nil {
		panic(fmt.Errorf("failed to analyze path: %w", err))
	}
	return info
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) GetQueryInfo() AnalyzedStruct {
	info, err := AnalyzeStruct(e.Query)
	if err != nil {
		panic(fmt.Errorf("failed to analyze query: %w", err))
	}
	return info
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) getHeaders() any {
	return e.Headers
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) ToPayload() (InputPayload, error) {

	bodyJsonBytes, err := json.Marshal(e.Body)
	if err != nil {
		return InputPayload{}, fmt.Errorf("failed to marshal body: %w", err)
	}

	serializeValue := func(value reflect.Value) (string, error) {
		str, err := json.Marshal(value.Interface())
		if err != nil {
			return "", fmt.Errorf("failed to marshal path parameter '%s': %w", value, err)
		}
		return strings.TrimSuffix(strings.TrimPrefix(string(str), "\""), "\""), nil
	}

	// Serialize headers
	headers := map[string][]string{}
	headersInfo, err := AnalyzeStruct(e.Headers)
	if err != nil {
		return InputPayload{}, fmt.Errorf("failed to analyze headers: %w", err)
	}
	for _, field := range headersInfo.Fields {
		if field.Name != "_" {
			key := field.LKName
			valuePtr, err := field.GetPtr(&e.Headers)
			if err != nil {
				return InputPayload{}, fmt.Errorf("failed to get header parameter '%s': %w", field.Name, err)
			}
			if valuePtr == nil && field.IsRequired() {
				return InputPayload{}, fmt.Errorf("missing required header parameter '%s'", field.Name)
			}
			if valuePtr == nil {
				continue
			}
			valueSerialized, err := serializeValue(reflect.ValueOf(valuePtr).Elem())
			if err != nil {
				return InputPayload{}, err
			}
			headers[key] = []string{valueSerialized}
		}
	}

	// serialize path
	path := map[string]string{}
	pathStr := ""
	pathT := reflect.TypeOf(e.Path)
	if pathT.Kind() != reflect.Struct {
		return InputPayload{}, fmt.Errorf("PathType must be a struct, this should have been caught in initial validation step")
	}
	for i := 0; i < pathT.NumField(); i++ {
		field := pathT.Field(i)
		if field.Name == "_" {
			// We won't bind this parameter, but it is still needed in the path
			// Check if it has a tag called path
			pathTag := field.Tag.Get("path")
			if pathTag == "" {
				// Treat as wildcard
				//path += "/*"
				return InputPayload{}, fmt.Errorf("wildcard path parameters not yet supported, field: %s", field.Name)
			} else {
				// Treat as literal
				pathStr += "/" + strings.TrimPrefix(pathTag, "/")
			}
		} else {
			valueSerialized, err := serializeValue(reflect.ValueOf(e.Path).Field(i))
			if err != nil {
				return InputPayload{}, err
			}
			pathStr += "/" + valueSerialized
			path[field.Name] = valueSerialized
		}
	}

	// Serialize query
	query := map[string][]string{}
	queryT := reflect.TypeOf(e.Query)
	if queryT.Kind() != reflect.Struct {
		return InputPayload{}, fmt.Errorf("QueryType must be a struct, this should have been caught in initial validation step")
	}
	for i := 0; i < queryT.NumField(); i++ {

		// if the value is nil, move on
		tpe := queryT.Field(i).Type
		value := reflect.ValueOf(e.Query).Field(i)
		if tpe.Kind() == reflect.Ptr && value.IsNil() {
			continue
		}

		field := queryT.Field(i)
		if field.Name != "_" {
			name := field.Name
			if nameOvrd, ok := field.Tag.Lookup("name"); ok {
				name = nameOvrd
			}
			valueSerialized, err := serializeValue(value)
			if err != nil {
				return InputPayload{}, err
			}
			query[name] = []string{valueSerialized}
		}
	}

	return InputPayload{
		Headers: headers,
		Path:    path,
		PathStr: pathStr,
		Query:   query,
		Body:    bodyJsonBytes,
	}, nil
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) getPath() any {
	return e.Path
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) parse(
	payload InputPayload,
	headerBindings HeaderBindings,
	pathBindings PathBindings,
	queryBindings QueryBindings,
) (any, error) {

	var result EndpointInput[HeadersType, PathType, QueryType, BodyType]

	// parse headers
	headerStructInfo, err := AnalyzeStruct(e.Headers)
	if err != nil {
		return result, fmt.Errorf("failed to analyze headers: %w", err)
	}
	for name, setter := range headerBindings.Bindings {
		lkName := strings.ToLower(name)
		fieldInfo, ok := headerStructInfo.FieldsByLKName[lkName]
		if !ok {
			return result, fmt.Errorf("failed to find header parameter info '%s'", name)
		}
		inputValue := payload.Headers[lkName]

		rootStructValue := reflect.ValueOf(&result.Headers).Elem()
		valueToSet := rootStructValue.FieldByName(fieldInfo.FieldName)
		reflectZero := reflect.Value{}
		if valueToSet == reflectZero {
			// Did not find the correct field. Try to find it by tag
			rootStructType := rootStructValue.Type()
			for i := 0; i < rootStructType.NumField(); i++ {
				field := rootStructType.Field(i)
				if strings.ToLower(field.Tag.Get("name")) == strings.ToLower(name) {
					valueToSet = rootStructValue.Field(i)
					break
				}
			}
		}

		if valueToSet == reflectZero {
			return result, fmt.Errorf("failed to find header parameter mapping for '%s'", name)
		}

		if len(inputValue) > 1 {
			return result, fmt.Errorf("repeated header parameters not yet supported, field: %s", name)
		}
		var err error
		if len(inputValue) == 0 {
			err = setter(valueToSet, nil)
		} else {
			err = setter(valueToSet, &inputValue[0])
		}
		if err != nil {
			return result, fmt.Errorf("failed to set header parameter '%s': %w", name, err)
		}
	}

	// parse path parameters
	for name, setter := range pathBindings.Bindings {
		inputValue, ok := payload.Path[name]
		if !ok {
			return result, fmt.Errorf("missing path parameter '%s'", name)
		}
		valueToSet := reflect.ValueOf(&result.Path).Elem().FieldByName(name)
		err := setter(valueToSet, inputValue)
		if err != nil {
			return result, fmt.Errorf("failed to set path parameter '%s': %w", name, err)
		}
	}

	// parse query parameters
	for name, setter := range queryBindings.Bindings {
		inputValue := payload.Query[name]
		valueToSet := reflect.ValueOf(&result.Query).Elem().FieldByName(name)
		if len(inputValue) > 1 {
			return result, fmt.Errorf("repeated query parameters not yet supported, field: %s", name)
		}
		var err error
		if len(inputValue) == 0 {
			err = setter(valueToSet, nil)
		} else {
			err = setter(valueToSet, &inputValue[0])
		}
		if err != nil {
			return result, fmt.Errorf("failed to set query parameter '%s': %w", name, err)
		}
	}

	// parse body
	bodyT := reflect.TypeOf(e.Body)
	if bodyT.Kind() != reflect.Struct {
		return result, NewError(http.StatusInternalServerError, "BodyType must be a struct, this should have been caught in initial validation step", nil)
	}

	// num fields in body
	numFields := bodyT.NumField()
	if numFields >= 1 {
		err := json.Unmarshal(payload.Body, &result.Body)
		if err != nil {
			return result, fmt.Errorf("failed to unmarshal body: %w", err)
		}
	}

	return result, nil
}

type HeaderBindings struct {
	Bindings map[string]headerFieldSetter
}

type PathBindings struct {
	FlatPath string
	Bindings map[string]pathFieldSetter
}

type QueryBindings struct {
	FlatPath string
	Bindings map[string]queryFieldSetter
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) calcHeaderBindings() HeaderBindings {

	structInfo, err := AnalyzeStruct(e.Headers)
	if err != nil {
		panic(fmt.Errorf("failed to analyze headers: %w", err))
	}

	result := HeaderBindings{
		Bindings: make(map[string]headerFieldSetter),
	}
	alreadyTaken := make(map[string]bool)

	// Iterate over fields in HeaderType
	for _, field := range structInfo.Fields {
		if field.Name != "_" {
			key := field.LKName
			if alreadyTaken[key] {
				panic(fmt.Sprintf("header '%s' is already taken", key))
			}
			alreadyTaken[key] = true
			result.Bindings[key] = getFromStringHeaderFieldSetter(field.StructField, key)
		}
	}

	return result
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) calcPathBindings() PathBindings {

	pathT := reflect.TypeOf((*PathType)(nil)).Elem()
	if pathT.Kind() != reflect.Struct {
		panic(fmt.Sprintf("PathType must be a struct, but is a %s", pathT.Kind().String()))
	}

	result := PathBindings{
		FlatPath: "",
		Bindings: make(map[string]pathFieldSetter),
	}
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
			result.Bindings[field.Name] = getFromStringPathFieldSetter(field)
		}
	}

	return result
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) calcQueryBindings() QueryBindings {

	pathT := reflect.TypeOf((*QueryType)(nil)).Elem()
	if pathT.Kind() != reflect.Struct {
		panic("QueryType must be a struct")
	}

	result := QueryBindings{
		Bindings: make(map[string]queryFieldSetter),
	}
	alreadyTaken := make(map[string]bool)

	// Iterate over fields in PathType
	for i := 0; i < pathT.NumField(); i++ {
		// Check if the field has path
		field := pathT.Field(i)

		if field.Name != "_" {
			if alreadyTaken[field.Name] {
				panic(fmt.Sprintf("field '%s' is already taken", field.Name))
			}

			isFirst := len(result.Bindings) == 0
			separator := "&"
			if isFirst {
				separator = "?"
			}
			alreadyTaken[field.Name] = true
			result.FlatPath += separator + field.Name + "=.."
			result.Bindings[field.Name] = getFromStringQueryFieldSetter(field)
		}
	}

	return result
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) validateBodyType() {
	bodyT := reflect.TypeOf(e.Body)
	if bodyT.Kind() != reflect.Struct {
		panic("BodyType must be a struct")
	}
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) getQuery() any {
	return e.Query
}

func (e EndpointInput[HeadersType, PathType, QueryType, BodyType]) getBody() any {
	return e.Body
}

func calcHeaderBindings[Input EndpointInputBase]() HeaderBindings {
	var zero Input
	return zero.calcHeaderBindings()
}

func calcPathBindings[Input EndpointInputBase]() PathBindings {
	var zero Input
	return zero.calcPathBindings()
}

func calcQueryBindings[Input EndpointInputBase]() QueryBindings {
	var zero Input
	return zero.calcQueryBindings()
}
