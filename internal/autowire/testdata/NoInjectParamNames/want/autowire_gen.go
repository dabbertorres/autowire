// Code generated by Autowire. DO NOT EDIT.

//go:generate go run github.com/dabbertorres/autowire/cmd/autowire

//go:build !wireinject
// +build !wireinject

package main

import (
	context2 "context"
)

// Injectors from autowire.go:

func inject(contextContext context2.Context, arg struct{}) (context, error) {
	mainContext, err := provide(contextContext)
	if err != nil {
		return context{}, err
	}
	return mainContext, nil
}
