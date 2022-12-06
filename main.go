package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var NewsQuantity int = 15
var SkipNews int = 0
var UntilDate int = 1666056885
var Arts []interface{}
var Coll *mongo.Collection
var News []Article

type TemplateData map[string]interface{}

func main() {
	// GET ARTICLES
	GetArticles(GetNewsList(NewsQuery(NewsQuantity, SkipNews, UntilDate)))
	// LOAD ENV
	godotenv.Load(".env")
	// TEMPLATE
	// Tmpl, err := template.ParseGlob("static/*")
	// if err != nil {
	// 	panic(err.Error())
	// }
	// MONGO CONNECTION
	MONGO_URL := os.Getenv("MONGO_URL")
	fmt.Println("Mongo URL = ", MONGO_URL)
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(MONGO_URL).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	Coll = client.Database("point").Collection("articles")

	// res, err := DeleteAllArticles()
	// if err != nil {
	// 	fmt.Println(res, err)
	// }
	// fmt.Println("Documents deleted: ", res)

	// err = InsertArticles(Arts)
	// if err != nil {
	// 	fmt.Println("Save to DB Error", err)
	// }

	ar := NewArticleRepo(Coll)
	err = ar.BulkWrite(News)
	if err != nil {
		fmt.Println("BulkWrite to DB Error", err)
	}

	//ROUTER
	server := NewServer(ar)
	router := httprouter.New()
	// static files
	router.ServeFiles("/static/*filepath", http.Dir("./static/"))
	// http.HandleFunc("/", Index)

	// dinamic router & API
	router.GET("/hello/:name", Hello)
	router.GET("/hi", Hi)
	router.POST("/form", Form)
	router.GET("/api/v1/article/:id", GetOneArticle)
	router.GET("/article/:id", GetOneArticlePage)
	router.GET("/api/v1/newslist", server.GetNews)
	router.GET("/news", server.GetNewsPage)
	router.GET("/search", server.Search)
	router.POST("/article/:id", EditArticle)

	// start server
	fmt.Println("Setrver start at port 8088")
	log.Fatal(http.ListenAndServe(":8088", router))
}

func (s *Server) Search(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tmpl, err := template.ParseFiles("static/pages/search.html", "static/partials/header.html", "static/partials/footer.html", "static/partials/head.html")
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	query := r.URL.Query().Get("query")
	td := TemplateData{"title": "Searching for: " + query}
	td["data"] = []Article{}
	fmt.Printf("func handler Search for query %s\n", query)
	// resp := &NewsRespons{}
	// resp.Message = "OK"
	// resp.Data = s.ar.GetNewsFromDB()
	// fmt.Println("func GetNewsPage, print first article ID ", resp.Data[0].Data.Content.Id)
	// fmt.Println(resp.Data[len(resp.Data)-1].Data.Content.Title.Short)
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// td["data"] = resp
	err = tmpl.ExecuteTemplate(w, "search", td)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Fatal(err)
	}
}

func EditArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Println("func EditArticle start")
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	title := r.PostForm.Get("title")
	description := r.PostForm.Get("description")
	id := ps.ByName("id")
	// title := r.FormValue("title")
	// description := r.FormValue("description")
	if title == "" || description == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "No title or description.\n Title: %s\n Description: %s", title, description)
		return
	}
	// fmt.Println("func Form data: ", id, title, description)
	art, ok := GetArticleById(id)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
	}
	art.Data.Content.Description.Long = description
	art.Data.Content.Title.Short = title
	err := UpdateOne(art)
	if err != nil {
		fmt.Println("func EditArticle => UpdateOne article error: ", err)
	}
	http.Redirect(w, r, "/article/"+id, http.StatusSeeOther)
}

func GetOneArticlePage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	res, ok := GetArticleById(id)
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
	tmpl, err := template.ParseFiles("static/pages/news.html", "static/partials/header.html", "static/partials/footer.html", "static/partials/head.html")
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	td := TemplateData{"title": "News, Analysis, Politics, Business, Technology"}
	fmt.Println("func handler GetNewsPage ")
	resp := &NewsRespons{}
	resp.Message = "OK"
	resp.Data = s.ar.GetNewsFromDB()
	fmt.Println("func GetNewsPage, print first article ID ", resp.Data[0].Data.Content.Id)
	fmt.Println(resp.Data[len(resp.Data)-1].Data.Content.Title.Short)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	td["data"] = resp
	err = tmpl.ExecuteTemplate(w, "news", td)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Fatal(err)
	}
}
