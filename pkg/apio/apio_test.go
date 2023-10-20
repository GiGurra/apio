package apio

import (
	"fmt"
	"reflect"
	"testing"
)

type UserSettingPath struct {
	_          any `path:"/users"`
	collection int
	_          any `path:"/settings"`
	category   string
	value      string
} // forms "/users/:collection/settings/:category/:value"

func TestGetUserSetting(t *testing.T) {
	input := EndpointInput[NoHeaders, UserSettingPath, NoQuery, NoBody]{}

	fmt.Printf("input: %v\n", reflect.TypeOf(input))
}
