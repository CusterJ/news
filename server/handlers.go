package server

import (
	"News/domain"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
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

	// check err form from last get if it was
	formError := r.URL.Query().Get("form")
	if formError == "error" {
		td["error"] = "Wrong name or password"
	}

	// check AUTH
	ac, ok := s.usecases.VerifyAuthCookies(r)
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

	ac, err := s.usecases.UserLogin(username, password, r.UserAgent())
	if err != nil {
		fmt.Println("func PostLogin handler -> UserLogin error -> Redirect: ", err)
		http.Redirect(w, r, "/login?form=error", http.StatusSeeOther)

		return
	}

	http.SetCookie(w, &ac)

	fmt.Printf("func PostLogin handler -> end \n User name: %s\n Password: %s\n", username, password)
	http.Redirect(w, r, "/news", http.StatusSeeOther)
}

func (s *Server) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Println("func Register handler -> start")

	if err := r.ParseForm(); err != nil {
		fmt.Println("func Register handler -> ParseForm error")

		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if len(username) < 3 || len(password) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Name or passwrod lengh is wrong.\n Name: %s - len is %v \n Password: %s - len is %v",
			username, len(username), password, len(password))

		return
	}

	ac, err := s.usecases.UserSave(username, password, r.UserAgent())
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
	fmt.Fprintf(w, "hello, %s!\n", name)
}

// REST API method for take paginated articles.
func (s *Server) GetNews(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pg := 0

	page := r.URL.Query().Get("page")
	if page != "" {
		page, err := strconv.Atoi(page)
		if err != nil {
			fmt.Fprint(w, "Bad page parameter", http.StatusBadRequest)

			return
		}

		pg = page
	}

	articleList, err := s.usecases.GetArticlesList(context.TODO(), pg)
	if err != nil {
		fmt.Fprintf(w, "Error while getting articless. Err: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	resp := &domain.ArticlesResponse{}
	resp.Message = "OK"
	resp.Data = articleList

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.Write(jsonResp)
}

// REST API method for get one aritcle by ID.
func (s *Server) GetOneArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	res, err := s.usecases.GetByID(r.Context(), id)
	if err != nil {
		fmt.Println("No id in db. Error: ", err)
		fmt.Fprint(w, err)
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
	tmpl, err := template.ParseFiles(
		"static/pages/search.html",
		"static/partials/header.html",
		"static/partials/footer.html",
		"static/partials/head.html",
		"static/partials/pagination.html",
	)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)

		return
	}

	query := r.URL.Query().Get("query")

	page := 1

	pg := r.URL.Query().Get("page")
	if pg != "" {
		page, err = strconv.Atoi(pg)
		if err != nil {
			http.Error(w, "Bad page parameter", http.StatusInternalServerError)

			return
		}
	} else {
		pg = "1"
	}

	arts, hits, err := s.usecases.Search(r.Context(), query, page)
	if err != nil {
		log.Fatal(err)
	}

	td := templateData{"title": "Searching for: " + query}
	td["data"] = arts
	td["hits"] = hits

	ac, ok := s.usecases.ReadAuthCookies(r)
	if ok {
		td["username"] = ac.Username
	}

	td["pagination"] = map[string]interface{}{"current": pg, "pages": Pagination(page, hits)}

	err = tmpl.ExecuteTemplate(w, "search", td)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Fatal(err)
	}
}

func (s *Server) EditArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	// Check if article with this ID exists in DB
	art, err := s.usecases.GetByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
	}
	// Replace Title and Description
	art.Title.Short = title
	art.Description.Long = description

	// Update article
	err = s.usecases.EditArticle(art)
	if err != nil {
		fmt.Println("func EditArticle => UpdateOne article error: ", err)
	}

	http.Redirect(w, r, "/article/"+id, http.StatusSeeOther)
}

func (s *Server) GetOneArticlePage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	edit := r.URL.Query().Get("edit")
	td := templateData{}

	tmpl, err := template.ParseFiles(
		"static/pages/article.html",
		"static/partials/header.html",
		"static/partials/footer.html",
		"static/partials/head.html",
	)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)

		return
	}

	ac, ok := s.usecases.ReadAuthCookies(r)

	if ok {
		td["username"] = ac.Username
	}

	if ok && edit == "true" {
		td["edit"] = true
	}

	id := ps.ByName("id")

	res, err := s.usecases.GetByID(r.Context(), id)
	if err != nil {
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
	tmpl, err := template.ParseFiles(
		"static/pages/news.html",
		"static/partials/header.html",
		"static/partials/footer.html",
		"static/partials/head.html",
		"static/partials/pagination.html",
	)
	if err != nil {
		http.Error(w, "Something with template went wrong", http.StatusInternalServerError)

		return
	}

	td := templateData{"title": "News, Analysis, Politics, Business, Technology"}

	ac, ok := s.usecases.ReadAuthCookies(r)
	if ok {
		td["username"] = ac.Username
	}

	page := 1

	pg := r.URL.Query().Get("page")
	if pg != "" {
		page, err = strconv.Atoi(pg)
		if err != nil {
			fmt.Fprint(w, "Bad page parameter", http.StatusBadRequest)

			return
		}
	} else {
		pg = "1"
	}

	arts, err := s.usecases.GetArticlesList(r.Context(), page)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)

		return
	}

	td["data"] = arts

	total, err := s.usecases.Count(r.Context())
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}

	pages := Pagination(page, int(total))
	pagination := map[string]interface{}{"current": pg, "pages": pages}
	td["pagination"] = pagination

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "news", td)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		fmt.Printf("func GetNewsPage -> ExecuteTemplate error: %v", err)
	}
}
