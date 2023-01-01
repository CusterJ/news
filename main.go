package main

import (
	"News/db/es"
	"News/db/mdb"
	"News/domain"
	newsparser "News/pkg/news_parser"
	"News/server"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PerPage int = 15

var News []domain.Article

type Articles struct {
	Arts  []domain.Article
	Total int
}

func main() {
	// GET ARTICLES
	newsparser.Start()
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
	Coll := client.Database("point").Collection("articles")
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
	server := server.NewServer(article)
	router := httprouter.New()
	// static files
	router.ServeFiles("/static/*filepath", http.Dir("./static/"))
	// http.HandleFunc("/", Index)

	// dinamic router & API
	router.GET("/hello/:name", server.Hello)
	router.GET("/hi", server.Hi)
	router.POST("/form", server.Form)
	router.GET("/api/v1/article/:id", server.GetOneArticle)
	router.GET("/article/:id", server.GetOneArticlePage)
	router.POST("/article/:id", server.EditArticle)
	router.GET("/api/v1/newslist", server.GetNews)
	router.GET("/news", server.GetNewsPage)
	router.GET("/search", server.Search)

	// start server
	fmt.Println("Setrver start at port 8088")
	log.Fatal(http.ListenAndServe(":8088", router))
}

func Check(err error) {
	if err != nil {
		log.Print(err)
	}
}
