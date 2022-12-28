package mdb

import (
	"News/domain"
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ArticleRepo struct {
	coll     *mongo.Collection
	Articles *domain.Article
}

func NewArticleRepo(coll *mongo.Collection) *ArticleRepo {
	return &ArticleRepo{coll: coll}
}

func (ar *ArticleRepo) ArticleList(ctx context.Context, in *domain.ArticlesRequest) (*domain.ArticlesResponse, error) {
	opts := options.Find().SetSkip(int64(in.Skip)).SetLimit(int64(in.Limit))

	cursor, err := ar.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}

	res := domain.ArticlesResponse{}
	err = cursor.All(ctx, &res.Data)
	if err != nil {
		return nil, err
	}

	res.Count, err = ar.coll.CountDocuments(ctx, bson.M{})

	return &res, err
}

// coll := client.Database("point").Collection("articles")

func (ar *ArticleRepo) InsertArticles(articles []interface{}) error {
	result, err := ar.coll.InsertMany(context.TODO(), articles)
	if err != nil {
		return err
	}
	fmt.Println("Inserted documents: ", len(result.InsertedIDs), result)
	return nil
}

func (ar *ArticleRepo) InsertOne(a domain.Article) error {
	result, err := ar.coll.InsertOne(context.TODO(), a)
	if err != nil {
		return err
	}
	fmt.Println("Inserted document: ", result, a.Data.Content.Title.Short)
	return nil
}

func (ar *ArticleRepo) FindAndInsert(a domain.Article) bool {
	fmt.Println("func FindAndInsert ", a.Data.Content.Id)
	_, ok := ar.GetArticleById(a.Data.Content.Id)
	if !ok {
		fmt.Println("Writing new document")
		ar.InsertOne(a)
		return false
	}
	fmt.Println("Document already exists")
	return true
}

func (ar *ArticleRepo) FindAndInsertMany(a []domain.Article) bool {
	inserted := 0
	fmt.Printf("func FindAndInsertMany from %v documents \n", len(a))
	for i := 0; i < len(a); i++ {
		_, ok := ar.GetArticleById(a[i].Data.Content.Id)
		if !ok {
			ar.InsertOne(a[i])
			inserted++
		}
	}
	fmt.Println("Inserted: ", inserted)
	return inserted <= 0
}

func (ar *ArticleRepo) GetArticleById(id string) (result domain.Article, ok bool) {
	filter := bson.M{"data.content.id": id}
	err := ar.coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		fmt.Printf("Can't find this Article %s in DB %v \n", id, err)
		return result, false
	}
	fmt.Printf("func GetArticleById document %s exists in DB\n", result.Data.Content.Title.Short)
	return result, true
}

func (ar *ArticleRepo) DeleteAllArticles() (string, error) {
	result, err := ar.coll.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		return "delete all documents ERROR", err
	}
	return fmt.Sprint(result), nil
}

func (ar *ArticleRepo) UpdateOne(a domain.Article) error {
	fmt.Println("UpdateOne start", a)
	opts := options.Update().SetUpsert(true)
	filter := bson.D{{Key: "data.content.id", Value: a.Data.Content.Id}}
	// filter := bson.D{{Key: "_id", Value: "415695b0-0dec-4d01-b8d7-e2c7ba7700ce"}}

	fmt.Println("Filter", filter)
	update := bson.D{{Key: "$set", Value: a}}
	// update := bson.D{{Key: "$set", Value: bson.D{{Key: "email", Value: "newemail@example.com"}}}}
	result, err := ar.coll.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}
	fmt.Printf("func Find one:\n Matched %v\n Modified %v\n Upserted %v\n UpsertedID %v\n",
		result.MatchedCount, result.ModifiedCount, result.UpsertedCount, result.UpsertedID)
	return nil
}

func (ar *ArticleRepo) BulkWrite(a []domain.Article) error {
	models := []mongo.WriteModel{}
	for i := 0; i < len(a); i++ {
		// fmt.Println(a[i])
		update := bson.D{{Key: "$set", Value: a[i]}}
		m := mongo.NewUpdateOneModel().SetFilter(bson.D{{Key: "data.content.id", Value: a[i].Data.Content.Id}}).SetUpdate(update).SetUpsert(true)
		models = append(models, m)
	}

	fmt.Println("BulkWrite func  :", models)
	opts := options.BulkWrite().SetOrdered(false)
	res, err := ar.coll.BulkWrite(context.TODO(), models, opts)
	if err != nil {
		return err
	}

	fmt.Printf("inserted %v\n upserted %v\n deleted %v\n matched %v\n modified %v\n upsertedIDs %v\n", res.InsertedCount, res.UpsertedCount, res.DeletedCount, res.MatchedCount, res.ModifiedCount, res.UpsertedIDs)
	return nil
}

func (ar *ArticleRepo) GetNewsFromDB() []domain.Article {
	// sort := desc
	var limit int64 = 15
	var skip int64 = 0
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(limit).SetSkip(skip)
	cursor, err := ar.coll.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		fmt.Println("func GetNewsFromDB error: ", err)
		return nil
	}
	var results []domain.Article
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	// for _, result := range results {
	// 	fmt.Println(result)
	// }
	fmt.Println("func GetNewsFromDB ends")
	return results
}
