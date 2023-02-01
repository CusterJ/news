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
	opts := options.Find().SetSort(bson.D{{Key: "dates.posted", Value: -1}}).SetSkip(int64(in.Skip)).SetLimit(int64(in.Limit))

	cursor, err := ar.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		fmt.Println("func ArticleList Find error: ", err)
		return nil, err
	}

	res := domain.ArticlesResponse{}
	err = cursor.All(ctx, &res.Data)
	if err != nil {
		fmt.Println("func ArticleList cursor All error: ", err)
		return nil, err
	}

	res.Count, err = ar.coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		fmt.Println("func ArticleList CountDocuments error: ", err)
	}

	return &res, err
}

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
	fmt.Println("Inserted document: ", result, a.Title.Short)
	return nil
}

func (ar *ArticleRepo) FindAndInsert(a domain.Article) bool {
	fmt.Println("func FindAndInsert ", a.Id)
	_, ok := ar.GetArticleById(a.Id)
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
		_, ok := ar.GetArticleById(a[i].Id)
		if !ok {
			ar.InsertOne(a[i])
			inserted++
		}
	}
	fmt.Println("Inserted: ", inserted)
	return inserted <= 0
}

func (ar *ArticleRepo) GetArticleById(id string) (result domain.Article, ok bool) {
	filter := bson.M{"id": id}
	err := ar.coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		fmt.Printf("Can't find this Article %s in DB %v \n", id, err)
		return result, false
	}
	fmt.Printf("func GetArticleById document %s exists in DB\n", result.Title.Short)
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
	filter := bson.D{{Key: "id", Value: a.Id}}
	update := bson.D{{Key: "$set", Value: a}}
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
		update := bson.D{{Key: "$set", Value: a[i]}}
		m := mongo.NewUpdateOneModel().SetFilter(bson.D{{Key: "id", Value: a[i].Id}}).SetUpdate(update).SetUpsert(true)
		models = append(models, m)
	}

	opts := options.BulkWrite().SetOrdered(false)
	res, err := ar.coll.BulkWrite(context.TODO(), models, opts)
	if err != nil {
		return err
	}

	fmt.Printf("Mongo: inserted %v, upserted %v, deleted %v, matched %v, modified %v\n",
		res.InsertedCount, res.UpsertedCount, res.DeletedCount, res.MatchedCount, res.ModifiedCount)
	return nil
}

func (ar *ArticleRepo) GetNewsFromDB(limit, skip int64) []domain.Article {
	fmt.Println("func GetNewsFromDB -> start")

	if limit == 0 {
		limit = 15
	}
	// sort := date desc
	opts := options.Find().SetSort(bson.D{{Key: "dates.posted", Value: -1}}).SetLimit(limit).SetSkip(skip)
	cursor, err := ar.coll.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		fmt.Println("func GetNewsFromDB error: ", err)
		return nil
	}
	var results []domain.Article
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}

	return results
}
