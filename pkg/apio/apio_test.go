package apio

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"log/slog"
	"net/http"
	"testing"
)

// represents "/users/:user/settings/:settingCat/:settingId"
type UserSettingPath struct {
	_          any `path:"/users"`
	User       int
	_          any `path:"/settings"`
	SettingCat string
	SettingId  string
}

type UserSettingQuery struct {
	Foo *string
	Bar int
}

type UserSetting struct {
	Value any     `json:"value"`
	Type  string  `json:"type"`
	Opt   *string `json:"opt"`
}

type UserSettingHeaders struct {
	Yo          any
	ContentType string `name:"Content-Type"`
}

type OutputHeaders struct {
	ContentType string `name:"Content-Type"`
}

func UserSettingEndpoints() []EndpointBase {

	return []EndpointBase{

		Endpoint[
			EndpointInput[UserSettingHeaders, UserSettingPath, UserSettingQuery, X],
			EndpointOutput[X, UserSetting],
		]{
			Method: http.MethodGet,
			Handler: func(
				input EndpointInput[UserSettingHeaders, UserSettingPath, UserSettingQuery, X],
			) (EndpointOutput[X, UserSetting], error) {
				slog.Info(fmt.Sprintf("invoked GET path with input: %+v", input))
				return BodyResponse(UserSetting{
					Value: "testValue",
					Type:  fmt.Sprintf("input=%+v", input),
				}), nil
			},
		},

		Endpoint[
			EndpointInput[X, UserSettingPath, X, UserSetting],
			EndpointOutput[OutputHeaders, X],
		]{
			Method: http.MethodPut,
			Handler: func(
				input EndpointInput[X, UserSettingPath, X, UserSetting],
			) (EndpointOutput[OutputHeaders, X], error) {
				slog.Info(fmt.Sprintf("invoked PUT path with input: %+v", input))
				return HeadersResponse(OutputHeaders{
					ContentType: "application/json",
				}), nil
			},
		},
	}
}

var testApi = Api{
	Name: "My test API",
	Servers: []Server{{
		Scheme:   "https",
		Host:     "api.example.com",
		Port:     443,
		BasePath: "/api/v1",
		HttpVer:  "1.1",
	}},
	IntBasePath: "/api/v1",
}.WithEndpoints(
	UserSettingEndpoints()...,
).Validate(false)

func TestGetUserSetting(t *testing.T) {

	getEndpoint := func() EndpointBase {
		for _, e := range testApi.Endpoints {
			if e.GetMethod() == http.MethodGet {
				return e
			}
		}
		panic("no GET endpoint found")
	}()

	result, err := getEndpoint.Handle(InputPayload{
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
			"Yo":           {"da"},
		},
		Path: map[string]string{
			"User":       "123",
			"SettingCat": "foo",
			"SettingId":  "bar",
		},
		Query: map[string][]string{
			"Foo": {"foo"},
			"Bar": {"123"},
		},
	})

	if err != nil {
		t.Fatal(fmt.Errorf("failed to Handle call: %w", err))
	}

	fmt.Printf("result: %+v\n", result)
}

func TestGetBadRequestMissingHeader(t *testing.T) {

	getEndpoint := func() EndpointBase {
		for _, e := range testApi.Endpoints {
			if e.GetMethod() == http.MethodGet {
				return e
			}
		}
		panic("no GET endpoint found")
	}()

	_, err := getEndpoint.Handle(InputPayload{
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Path: map[string]string{
			"User":       "123",
			"SettingCat": "foo",
			"SettingId":  "bar",
		},
		Query: map[string][]string{
			"Foo": {"foo"},
			"Bar": {"123"},
		},
	})

	if err == nil {
		t.Fatal(fmt.Errorf("expected error with 400 out, but didnt get any"))
	}

	errTyped := *AsErResp(err)
	if errTyped.Status != http.StatusBadRequest {
		t.Fatal(fmt.Errorf("unexpected status code: %d", errTyped.Status))
	}
}

func TestGetUserMissingPath(t *testing.T) {

	getEndpoint := func() EndpointBase {
		for _, e := range testApi.Endpoints {
			if e.GetMethod() == http.MethodGet {
				return e
			}
		}
		panic("no GET endpoint found")
	}()

	_, err := getEndpoint.Handle(InputPayload{
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
			"Yo":           {"da"},
		},
		Path: map[string]string{
			"User":       "123",
			"SettingCat": "foo",
		},
		Query: map[string][]string{
			"Foo": {"foo"},
			"Bar": {"123"},
		},
	})

	if err == nil {
		t.Fatal(fmt.Errorf("expected error with 400 out, but didnt get any"))
	}

	errTyped := *AsErResp(err)

	if errTyped.Status != http.StatusBadRequest {
		t.Fatal(fmt.Errorf("unexpected status code: %d", errTyped.Status))
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func TestPutUserSetting(t *testing.T) {

	getEndpoint := func() EndpointBase {
		for _, e := range testApi.Endpoints {
			if e.GetMethod() == http.MethodPut {
				return e
			}
		}
		panic("no PUT endpoint found")
	}()

	result, err := getEndpoint.Handle(InputPayload{
		Path: map[string]string{
			"User":       "123",
			"SettingCat": "foo",
			"SettingId":  "bar",
		},
		Body: must(json.Marshal(UserSetting{
			Value: "testValue",
			Type:  "testType",
		})),
	})

	if err != nil {
		t.Fatal(fmt.Errorf("failed to Handle call: %w", err))
	}

	expHeaders := map[string][]string{
		"Content-Type": {"application/json"},
	}

	if diff := cmp.Diff(result.GetHeaders(), expHeaders); diff != "" {
		t.Fatal(fmt.Errorf("unexpected headers: %s", diff))
	}

	fmt.Printf("result: %+v\n", result)
}
