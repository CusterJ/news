package server

import (
	"News/graph"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/julienschmidt/httprouter"
)

func (s *Server) PlaygroundHandler() httprouter.Handle {
	ph := playground.Handler("My GraphQL playground", "/query")

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ph(w, r)
	}
}

func (s *Server) GraphqlHandler(resolver *graph.Resolver) httprouter.Handle {
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers:  resolver,
		Directives: graph.DirectiveRoot{},
		Complexity: graph.ComplexityRoot{},
	}))

	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		srv.ServeHTTP(w, req)
	}
}
