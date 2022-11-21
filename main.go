package main

import (
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

var NewsQuantity int = 15
var SkipNews int = 0
var UntilDate int = 1666056885
var Arts []interface{}
var Coll *mongo.Collection
var News []Article

func main() {
	GetArticles(GetNewsList(NewsQuery(NewsQuantity, SkipNews, UntilDate)))
	godotenv.Load(".env")
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
	server := NewServer(ar)
	//ROUTER
	router := httprouter.New()
	// static files
	router.ServeFiles("/static/*filepath", http.Dir("./static/"))
	// http.HandleFunc("/", Index)

	// dinamic router & API
	router.GET("/hello/:name", Hello)
	router.GET("/hi", Hi)
	router.POST("/form", Form)
	router.GET("/api/v1/article/:id", GetOneArticle)
	router.GET("/api/v1/newslist", server.GetNews)

	// start server
	fmt.Println("Setrver start at port 8088")
	log.Fatal(http.ListenAndServe(":8088", router))
}
