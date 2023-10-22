package main

import (
	"fmt"
	"github.com/GiGurra/apio/cmd/examples/user_setting"
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

	input1 := NewInput(
		user_setting.Headers{
			Yo:          "yo",
			ContentType: "application/json",
		},
		user_setting.PathById{
			User:       123,
			SettingCat: "cat",
			SettingId:  "id",
		},
		user_setting.Query{
			Foo: ptr("foo"),
			Bar: 123,
		},
		Empty, // no body in this get call
	)

	res1, err1 := user_setting.GetById.RPC(server, input1, DefaultOpts())
	if err1 != nil {
		panic(fmt.Sprintf("failed to call RPC GET endpoint: %v", err1))
	}

	fmt.Printf("res: %+v\n", res1)

	input2 := NewInput(
		user_setting.Headers{
			Yo:          "yo",
			ContentType: "application/json",
		},
		user_setting.PathAll{
			User: 123,
		},
		Empty, // no query in this get call
		Empty, // no body in this get call
	)

	res2, err2 := user_setting.GetAll.RPC(server, input2, DefaultOpts())
	if err2 != nil {
		panic(fmt.Sprintf("failed to call RPC GET all endpoint: %v", err1))
	}

	fmt.Printf("res: %+v\n", res2)
}
