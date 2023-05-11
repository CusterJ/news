package main

import (
	"News/db/es"
	"News/db/mdb"
	"News/graph"
	newsparser "News/pkg/news_parser"
	"News/server"
	"News/usecases"
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

func main() {
	godotenv.Load(".env")

	// MONGO CONNECTION
	mongoURL := os.Getenv("MONGO_URL")
	timeoutSeconds := 10
	timeoutDuration := time.Duration(timeoutSeconds) * time.Second
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(mongoURL).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)

	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Println("Mongo Connect error", err)
	}
	// MONGO COLLECTIONS
	database := client.Database("point")

	articlesColl := database.Collection("articles")
	articlesRepo := mdb.NewArticleRepo(articlesColl)

	usersColl := database.Collection("users")
	usersRepo := mdb.NewUserRepo(usersColl)

	// ELASTIC
	esArts := os.Getenv("ES_ARTS")
	searchRepo := es.NewElasticRepo(esArts)

	// NEW SERVER
	usecases := usecases.NewUseCases(articlesRepo, usersRepo, searchRepo)
	resolver := graph.NewGqlResolver(usecases)
	server := server.NewServer(usersRepo, usecases)

	// GET ARTICLES WITH PARSER
	parser := newsparser.NewWorker(articlesRepo, searchRepo)
	stopChan := make(chan bool)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go parser.StartParser(stopChan, wg)

	// ROUTER
	router := httprouter.New()
	// static files
	router.ServeFiles("/static/*filepath", http.Dir("./static/"))

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

	// GraphQL
	router.GET("/", server.PlaygroundHandler())
	router.POST("/query", server.GraphqlHandler(resolver))

	// start server
	fmt.Println("Setrver starting at port 8088")

	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	httpServer := http.Server{
		Addr:    ":8088",
		Handler: router,
	}

	httpServerShutdown(httpServer, gracefulShutdown)

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe Error: %v", err)
	}

	<-gracefulShutdown
	stopChan <- true

	wg.Wait()
	fmt.Println("Shutdown gracefully")
}

func httpServerShutdown(httpServer http.Server, gracefulShutdown chan os.Signal) {
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}

		close(gracefulShutdown)
	}()
}
