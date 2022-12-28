package main

import (
	"News/db/es"
	"News/db/mdb"
	"News/domain"
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PerPage int = 15

var NewsQuantity int = 15
var SkipNews int = 0
var UntilDate int = 1666056885
var Coll *mongo.Collection
var News []domain.Article

type Articles struct {
	Arts  []domain.Article
	Total int
}

type TemplateData map[string]interface{}

func main() {
	// GET ARTICLES
	GetArticles(GetNewsList(NewsQuery(NewsQuantity, SkipNews, UntilDate)))
	// LOAD ENV
	godotenv.Load(".env")
	MONGO_URL := os.Getenv("MONGO_URL")
	// ES_ARTS := os.Getenv("ES_ARTS")

	// MONGO CONNECTION
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
	article := mdb.NewArticleRepo(Coll)
	// res, err := DeleteAllArticles()
	// if err != nil {
	// 	fmt.Println(res, err)
	// }
	// fmt.Println("Documents deleted: ", res)

	// err = InsertArticles(Arts)
	// if err != nil {
	// 	fmt.Println("Save to DB Error", err)
	// }

	err = article.BulkWrite(News)
	if err != nil {
		fmt.Println("BulkWrite to DB Error", err)
	}

	// Elastic
	err = es.EsInsertBulk(News)
	if err != nil {
		fmt.Println("func EsInsertBulk error")
	}

	//ROUTER
	server := NewServer(article)
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
	router.POST("/article/:id", EditArticle)
	router.GET("/api/v1/newslist", server.GetNews)
	router.GET("/news", server.GetNewsPage)
	router.GET("/search", server.Search)

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

func EditArticle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	article := mdb.NewArticleRepo(Coll)
	art, ok := article.GetArticleById(id)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
	}
	art.Data.Content.Description.Long = description
	art.Data.Content.Title.Short = title
	err := article.UpdateOne(art)
	if err != nil {
		fmt.Println("func EditArticle => UpdateOne article error: ", err)
	}
	err = es.EsUpdateOne(art)
	Check(err)
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

	article := mdb.NewArticleRepo(Coll)
	res, ok := article.GetArticleById(id)
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
	Check(err)
	err = tmpl.ExecuteTemplate(w, "news", td)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		fmt.Printf("func GetNewsPage -> ExecuteTemplate error: %v", err)
	}
}

func Pagination(page, total int) (pagelist []string) {
	var take, allPages, skip int
	take, _ = strconv.Atoi(os.Getenv("TAKE"))
	allPages = int(math.Ceil(float64(total) / float64(take)))
	// fmt.Printf("total/ take type is %T \n %v / %v = %v \n", total/take, total, take, total/take)
	if page > 0 {
		skip = (page - 1) * take
	}
	if skip > total {
		skip = 0
	}
	for i := 1; i <= allPages; i++ {
		pagelist = append(pagelist, strconv.Itoa(i))
	}
	// fmt.Printf("total= %v, take = %v, page = %v, allPages = %v, skip = %v \n", total, take, page, allPages, skip)
	return
}

func Check(err error) {
	if err != nil {
		log.Print(err)
	}
}
