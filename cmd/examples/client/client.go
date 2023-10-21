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

	input := NewInput(
		user_setting.Headers{
			Yo:          "yo",
			ContentType: "application/json",
		},
		user_setting.Path{
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

	res, err := user_setting.Get.RPC(server, input, DefaultOpts())
	if err != nil {
		panic(fmt.Sprintf("failed to call RPC GET endpoint: %v", err))
	}

	fmt.Printf("res: %+v\n", res)
}
