package main

import (
	"fmt"
	. "github.com/GiGurra/apio/cmd/examples/common"
	. "github.com/GiGurra/apio/pkg/apio"
)

func ptr[T any](v T) *T {
	return &v
}

func main() {
	server := Server{
		Scheme:   "http",
		Host:     "localhost",
		Port:     8080,
		BasePath: "/api/v1",
		HttpVer:  "1.1",
	}

	input := NewInput(
		UserSettingHeaders{
			Yo:          "yo",
			ContentType: "application/json",
		},
		UserSettingPath{
			User:       123,
			SettingCat: "cat",
			SettingId:  "id",
		},
		UserSettingQuery{
			Foo: ptr("foo"),
			Bar: 123,
		},
		Empty, // no body in this get call
	)

	res, err := GetEndpointSpec.Call(server, input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("res: %+v\n", res)
}
