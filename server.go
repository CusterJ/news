package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Server struct {
	ar *ArticleRepo
}

func NewServer(a *ArticleRepo) *Server {
	return &Server{ar: a}
}

func Form(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Println("func Form start")
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	name := r.FormValue("name")
	age := r.FormValue("age")
	if name == "" || age == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "No name or age.\n Name: %s\n Age: %s", name, age)
		return
	}

	fmt.Fprintf(w, "POST request successful\n")
	fmt.Fprintf(w, "Name = %s\n", name)
	fmt.Fprintf(w, "Age = %s\n", age)
	fmt.Println("func Form data: ", name, age)
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
	for _, v := range ps {
		fmt.Println(v)
	}
}

func Hi(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := r.URL.Query().Get("name")
	fmt.Println("name is =>", name)
	fmt.Fprintf(w, "hello, %s!\n", name)

}

func (s *Server) GetNews(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := &NewsRespons{}
	resp.Message = "OK"
	resp.Data = s.ar.GetNewsFromDB()

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.Write(jsonResp)
}

func GetOneArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	fmt.Println(id)
	res, ok := GetArticleById(id)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := &ArticleRespons{}
	resp.Message = "OK"
	resp.Data = res
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}
