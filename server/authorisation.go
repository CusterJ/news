package server

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (s *Server) Protected(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		_, ok := s.usecases.VerifyAuthCookies(r)

		if !ok {
			fmt.Println("Check Auth - cookie error")
			http.Error(w, "cookie error -> FORBIDDEN", http.StatusForbidden)
			return
		}
		h(w, r, ps)
	}
}
