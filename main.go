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

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PerPage int = 15
const defaultGqlPort = "8080"

func main() {
	// LOAD ENV
	godotenv.Load(".env")
	MONGO_URL := os.Getenv("MONGO_URL")

	gqlPort := os.Getenv("GQL_PORT")
	if gqlPort == "" {
		gqlPort = defaultGqlPort
	}

	// MONGO CONNECTION
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(MONGO_URL).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	ES_ARTS := os.Getenv("ES_ARTS")
	searchRepo := es.NewElasticRepo(ES_ARTS)

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

	//ROUTER
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
	router.GET("/", playgroundHandler())
	router.POST("/query", GraphqlHandler(resolver))

	// start server
	fmt.Println("Setrver starting at port 8088")

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

func playgroundHandler() httprouter.Handle {

	ph := playground.Handler("My GraphQL playground", "/query")

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ph(w, r)
	}
}

func GraphqlHandler(resolver *graph.Resolver) httprouter.Handle {

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers:  resolver,
		Directives: graph.DirectiveRoot{},
		Complexity: graph.ComplexityRoot{},
	}))

	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		srv.ServeHTTP(w, req)
	}
}
