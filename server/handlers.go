package server

import (
	"News/domain"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
)

type templateData map[string]interface{}

func (s *Server) GetLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Printf("func handler GetLogin \n")

	tmpl, err := template.ParseFiles("static/pages/login.html")
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	td := templateData{}

	// check AUTH
	ac, ok := s.VerifyAuthCookies(r)
	if ok {
		td["auth"] = ac.Username
	}

	fmt.Println("func handler GetLogin -> Template Data -> AUTH check", td)

	err = tmpl.ExecuteTemplate(w, "login", td)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Fatal(err)
	}
}

func (s *Server) PostLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Println("func PostLogin handler -> start")

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "func PostLogin handler -> ParseForm() err: %v", err)
	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	// Return error
	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "User name or password is empty.\n Name: %s\n Password: %s", username, password)
		td := templateData{"error": bson.M{"login": "enter login", "password": "empty password"}}
		b, err := json.Marshal(td)
		if err != nil {
			fmt.Println("func PostLogin handler -> Marshal Template Data error")
		}
		w.Write(b)
		http.Redirect(w, r, "/login", http.StatusBadRequest)
	}

	ac, err := s.UserLogin(username, password, r.UserAgent())
	if err != nil {
		fmt.Println("func PostLogin handler -> UserLogin error -> Redirect: ", err)
		http.Redirect(w, r, "/login", http.StatusNoContent)
		return
	}

	http.SetCookie(w, &ac)

	fmt.Printf("func PostLogin handler -> end \n User name: %s\n Password: %s\n", username, password)
	http.Redirect(w, r, "/news", http.StatusSeeOther)

}

func (s *Server) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Println("func Register handler -> start")

	err := r.ParseForm()
	if err != nil {
		fmt.Println("func Register handler -> ParseForm error")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if len(username) < 3 || len(password) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Name or passwrod lengh is wrong.\n Name: %s - len is %v \n Password: %s - len is %v", username, len(username), password, len(password))
		return
	}

	ac, err := s.UserSave(username, password, r.UserAgent())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Create new User error.\n Name: %s\n Age: %s", username, password)
		return
	}

	http.SetCookie(w, &ac)
	fmt.Printf("Success! User created! User: %v => Password: %v\n", username, password)
	http.Redirect(w, r, "/news", http.StatusSeeOther)
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
	resp := &domain.ArticlesResponse{}
	resp.Message = "OK"
	resp.Data = s.ar.GetNewsFromDB(15, 0)

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

	td := templateData{"title": "Searching for: " + query}

	td["data"], err = s.es.EsSearchArticle(query)
	if err != nil {
		log.Fatal(err)
	}

	ac, ok := s.ReadAuthCookies(r)
	if ok {
		td["username"] = ac.Username
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
	art.Description.Long = description
	art.Title.Short = title
	err := s.ar.UpdateOne(art)
	if err != nil {
		fmt.Println("func EditArticle => UpdateOne article error: ", err)
	}
	err = s.es.EsUpdateOne(art)
	if err != nil {
		fmt.Println("EditArticle handler error -> EsUpdateOne error")
	}
	http.Redirect(w, r, "/article/"+id, http.StatusSeeOther)
}

func (s *Server) GetOneArticlePage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	edit := r.URL.Query().Get("edit")
	td := templateData{}
	tmpl, err := template.ParseFiles("static/pages/article.html", "static/partials/header.html", "static/partials/footer.html", "static/partials/head.html")
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	ac, ok := s.ReadAuthCookies(r)

	if ok {
		td["username"] = ac.Username
	}

	if ok && edit == "true" {
		td["edit"] = true
	}
	id := ps.ByName("id")
	fmt.Println("func handler GetOneArticlePage with id: ", id)

	res, ok := s.ar.GetArticleById(id)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	td["content"] = res

	date := time.Unix(res.Dates.Posted, 0)
	td["date"] = date

	err = tmpl.ExecuteTemplate(w, "article", td)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) GetNewsPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Println("func handler GetNewsPage ")

	tmpl, err := template.ParseFiles("static/pages/news.html", "static/partials/header.html", "static/partials/footer.html",
		"static/partials/head.html", "static/partials/pagination.html")
	if err != nil {
		http.Error(w, "Something with template went wrong", http.StatusInternalServerError)
		return
	}

	td := templateData{"title": "News, Analysis, Politics, Business, Technology"}

	ac, ok := s.ReadAuthCookies(r)
	if ok {
		td["username"] = ac.Username
	}

	// work with page query and skip for db req
	currentPage := 0

	cpage := r.URL.Query().Get("page")
	if cpage != "" {
		currentPage, _ = strconv.Atoi(cpage)
	}
	take, err := strconv.Atoi(os.Getenv("TAKE"))
	if err != nil {
		take = 15 // set default
	}

	skip := 0
	if currentPage > 1 {
		skip = (currentPage - 1) * take
	}

	fmt.Printf("GetNewsPage -> take: %v, page: %v, skip: %v \n", take, currentPage, skip)

	rd, err := s.ar.ArticleList(r.Context(), &domain.ArticlesRequest{
		Skip:  skip,
		Limit: take,
	})
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	td["data"] = rd.Data

	pages := Pagination(currentPage, int(rd.Count))
	pg := map[string]interface{}{"current": cpage, "pages": pages}
	td["pagination"] = pg

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "news", td)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		fmt.Printf("func GetNewsPage -> ExecuteTemplate error: %v", err)
	}
}
