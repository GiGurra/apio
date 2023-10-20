package apio

import "testing"

type ExamplePath struct {
	_          any `path:"/users"`
	collection int
	_          any `path:"/settings"`
	category   string
	value      string
} // forms "/users/:collection/settings/:category/:value"

func TestFakeMain(t *testing.T) {

}
