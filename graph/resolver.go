package graph

//go:generate go run github.com/99designs/gqlgen generate

import (
	"News/graph/model"
	"News/server"
	"News/usecases"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	todos    []*model.Todo
	srv      *server.Server
	usecases *usecases.UseCases
}

func NewGqlResolver(srv *server.Server, uc *usecases.UseCases) *Resolver {
	return &Resolver{
		todos:    []*model.Todo{},
		srv:      srv,
		usecases: uc,
	}
}
