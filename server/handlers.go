package server

import (
	"News/db/es"
	"News/domain"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type TemplateData map[string]interface{}

// Marshal API writer Newslist struct
type NewsRespons struct {
	Message string           `json:"message,omitempty"`
	Data    []domain.Article `json:"data,omitempty"`
}

func (s *Server) Form(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

func (s *Server) Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
	for _, v := range ps {
		fmt.Println(v)
	}
}

func (s *Server) Hi(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func (s *Server) GetOneArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	fmt.Println(id)
	res, ok := s.ar.GetArticleById(id)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := &domain.ArticleResponse{}
	resp.Message = "OK"
	resp.Data = res
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func (s *Server) Search(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tmpl, err := template.ParseFiles("static/pages/search.html", "static/partials/header.html", "static/partials/footer.html", "static/partials/head.html")
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	query := r.URL.Query().Get("query")
	fmt.Printf("func handler Search for query %s\n", query)
	td := TemplateData{"title": "Searching for: " + query}
	td["data"], err = es.EsSearchArticle(query)
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.ExecuteTemplate(w, "search", td)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Fatal(err)
	}
}

func (s *Server) EditArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Println("func EditArticle start")
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	title := r.PostForm.Get("title")
	description := r.PostForm.Get("description")
	id := ps.ByName("id")
	if title == "" || description == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "No title or description.\n Title: %s\n Description: %s", title, description)
		return
	}
	art, ok := s.ar.GetArticleById(id)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
	}
	art.Data.Content.Description.Long = description
	art.Data.Content.Title.Short = title
	err := s.ar.UpdateOne(art)
	if err != nil {
		fmt.Println("func EditArticle => UpdateOne article error: ", err)
	}
	err = es.EsUpdateOne(art)
	if err != nil {
		fmt.Println("EditArticle handler error -> EsUpdateOne error")
	}
	http.Redirect(w, r, "/article/"+id, http.StatusSeeOther)
}

func (s *Server) GetOneArticlePage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	edit := r.URL.Query().Get("edit")
	td := TemplateData{}
	tmpl, err := template.ParseFiles("static/pages/article.html", "static/partials/header.html", "static/partials/footer.html", "static/partials/head.html")
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if edit == "true" {
		td["edit"] = "true"
	}
	id := ps.ByName("id")
	fmt.Println("func handler GetOneArticlePage with id: ", id)

	res, ok := s.ar.GetArticleById(id)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	td["content"] = res.Data.Content
	err = tmpl.ExecuteTemplate(w, "article", td)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) GetNewsPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tmpl, err := template.ParseFiles("static/pages/news.html", "static/partials/header.html", "static/partials/footer.html",
		"static/partials/head.html", "static/partials/pagination.html")
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	currentPage := 0

	cpage := r.URL.Query().Get("page")
	if cpage != "" {
		currentPage, _ = strconv.Atoi(cpage)
	}
	take, _ := strconv.Atoi(os.Getenv("TAKE"))

	skip := 0
	if currentPage > 1 {
		skip = (currentPage - 1) * take
	}

	fmt.Printf("GetNewsPage -> take: %v, page: %v, skip: %v \n", take, currentPage, skip)

	td := TemplateData{"title": "News, Analysis, Politics, Business, Technology"}
	fmt.Println("func handler GetNewsPage ")
	resp := &NewsRespons{}
	resp.Message = "OK"

	rd, err := s.ar.ArticleList(r.Context(), &domain.ArticlesRequest{
		Skip:  skip,
		Limit: take,
	})
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	resp.Data = rd.Data

	pages := Pagination(currentPage, int(rd.Count))
	pg := map[string]interface{}{"current": cpage, "pages": pages}
	td["pagination"] = pg

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	td["data"] = resp
	err = tmpl.ExecuteTemplate(w, "news", td)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		fmt.Printf("func GetNewsPage -> ExecuteTemplate error: %v", err)
	}
}
