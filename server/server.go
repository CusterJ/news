package server

import (
	"News/db/es"
	"News/db/mdb"
)

type Server struct {
	ar *mdb.ArticleRepo
	ur *mdb.UserRepo
	es *es.ElasticRepo
}

func NewServer(ar *mdb.ArticleRepo, ur *mdb.UserRepo, es *es.ElasticRepo) *Server {
	// return &Server{ar: ar, es: es}
	return &Server{
		ar: ar,
		ur: ur,
		es: es,
	}
}

// func NewEsIndex(er *es.ElasticRepo) *Server {
// 	return &Server{es: er}
// }

// func serverMongoArticlesConnect() *Server {
// 	// LOAD ENV
// 	godotenv.Load(".env")
// 	MONGO_URL := os.Getenv("MONGO_URL")
// 	// ES_ARTS := os.Getenv("ES_ARTS")

// 	// MONGO CONNECTION
// 	fmt.Println("Mongo URL = ", MONGO_URL)
// 	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
// 	clientOptions := options.Client().
// 		ApplyURI(MONGO_URL).
// 		SetServerAPIOptions(serverAPIOptions)
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	client, err := mongo.Connect(ctx, clientOptions)
// 	if err != nil {
// 		fmt.Println("func server.MongoArticles mongo.Connect ERROR")
// 	}

// 	// MONGO COLLECTION
// 	articlesColl := client.Database("point").Collection("articles")
// 	articles := mdb.NewArticleRepo(articlesColl)

// 	server := NewArticleRepo(articles)
// 	return server
// }
