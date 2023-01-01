package server

import (
	"News/db/es"
	"News/db/mdb"
)

type Server struct {
	ar *mdb.ArticleRepo
	es *es.ElasticRepo
}

func NewServer(ar *mdb.ArticleRepo) *Server {
	// return &Server{ar: ar, es: es}
	return &Server{ar: ar}
}

func NewEsIndex(er *es.ElasticRepo) *Server {
	return &Server{es: er}
}
