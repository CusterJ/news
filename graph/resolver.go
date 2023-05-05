package graph

//go:generate go run github.com/99designs/gqlgen generate

import (
	"News/usecases"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	usecases *usecases.UseCases
}

func NewGqlResolver(uc *usecases.UseCases) *Resolver {
	return &Resolver{
		usecases: uc,
	}
}
