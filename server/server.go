package server

import (
	"News/db/mdb"
	"News/usecases"
)

type Server struct {
	ur       *mdb.UserRepo
	usecases *usecases.UseCases
}

func NewServer(ur *mdb.UserRepo, uc *usecases.UseCases) *Server {
	return &Server{
		ur:       ur,
		usecases: uc,
	}
}
