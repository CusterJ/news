package server

import (
	"News/db/es"
	"News/db/mdb"
)

type Server struct {
	ar *mdb.ArticleRepo
	ur *mdb.UserRepo
	es *es.ElasticRepo
}

func NewServer(ar *mdb.ArticleRepo, ur *mdb.UserRepo, es *es.ElasticRepo) *Server {
	// return &Server{ar: ar, es: es}
	return &Server{
		ar: ar,
		ur: ur,
		es: es,
	}
}
