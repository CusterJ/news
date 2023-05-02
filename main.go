package main

import (
	"News/db/es"
	"News/db/mdb"
	newsparser "News/pkg/news_parser"
	"News/server"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PerPage int = 15

func main() {
	// LOAD ENV
	godotenv.Load(".env")
	MONGO_URL := os.Getenv("MONGO_URL")

	// MONGO CONNECTION
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(MONGO_URL).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	Check(err)

	// MONGO COLLECTIONS
	database := client.Database("point")

	articlesColl := database.Collection("articles")
	articles := mdb.NewArticleRepo(articlesColl)

	usersColl := database.Collection("users")
	users := mdb.NewUserRepo(usersColl)

	// ELASTIC
	ES_ARTS := os.Getenv("ES_ARTS")
	esArticles := es.NewElasticRepo(ES_ARTS)

	// NEW SERVER
	server := server.NewServer(articles, users, esArticles)

	// GET ARTICLES WITH PARSER
	parser := newsparser.NewWorker(articles, esArticles)
	stopChan := make(chan bool)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go parser.StartParser(stopChan, wg)

	//ROUTER
	router := httprouter.New()
	// static files
	router.ServeFiles("/static/*filepath", http.Dir("./static/"))
	// http.HandleFunc("/", Index)

	// dinamic router & API
	router.GET("/hello/:name", server.Hello)
	router.GET("/hi", server.Hi)
	router.POST("/form", server.Form)
	router.GET("/article/:id", server.GetOneArticlePage)
	router.POST("/article/:id", server.Protected(server.EditArticle))
	router.GET("/api/v1/newslist", server.Protected(server.GetNews))
	router.GET("/api/v1/article/:id", server.Protected(server.GetOneArticle))
	router.GET("/news", server.GetNewsPage)
	router.GET("/search", server.Search)
	router.GET("/login", server.GetLogin)
	router.POST("/login", server.PostLogin)
	router.POST("/register", server.Register)

	// start server
	fmt.Println("Setrver start at port 8088")
	// log.Fatal(http.ListenAndServe(":8088", router))

	// ==============================
	// gracefull shutdown http server
	// ==============================
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	httpServer := http.Server{
		Addr:    ":8088",
		Handler: router,
	}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}

		close(gracefulShutdown)
	}()

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe Error: %v", err)
	}

	<-gracefulShutdown
	// ==============================
	stopChan <- true
	wg.Wait()
	// defer time.Sleep(3 * time.Second)
	fmt.Println("Shutdown gracefully")
}

func Check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
