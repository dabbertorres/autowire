// Code generated by Autowire. DO NOT EDIT.

//go:generate go run github.com/dabbertorres/autowire/cmd/autowire

//go:build !wireinject
// +build !wireinject

package main

// Injectors from autowire.go:

func injectFoo() Foo {
	foo := _wireFooValue
	return foo
}

var (
	_wireFooValue = Foo{X: 42}
)
