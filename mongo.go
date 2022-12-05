package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ArticleRepo struct {
	coll *mongo.Collection
}

func NewArticleRepo(c *mongo.Collection) *ArticleRepo {
	return &ArticleRepo{coll: c}
}

// coll := client.Database("point").Collection("articles")

func InsertArticles(articles []interface{}) error {
	result, err := Coll.InsertMany(context.TODO(), articles)
	if err != nil {
		return err
	}
	fmt.Println("Inserted documents: ", len(result.InsertedIDs), result)
	return nil
}

func InsertOne(a Article) error {
	result, err := Coll.InsertOne(context.TODO(), a)
	if err != nil {
		return err
	}
	fmt.Println("Inserted document: ", result, a.Data.Content.Title.Short)
	return nil
}

func FindAndInsert(a Article) bool {
	fmt.Println("func FindAndInsert ", a.Data.Content.Id)
	_, ok := GetArticleById(a.Data.Content.Id)
	if !ok {
		fmt.Println("Writing new document")
		InsertOne(a)
		return false
	}
	fmt.Println("Document already exists")
	return true
}

func FindAndInsertMany(a []Article) bool {
	inserted := 0
	fmt.Printf("func FindAndInsertMany from %v documents \n", len(a))
	for i := 0; i < len(a); i++ {
		_, ok := GetArticleById(a[i].Data.Content.Id)
		if !ok {
			InsertOne(a[i])
			inserted++
		}
	}
	fmt.Println("Inserted: ", inserted)
	return inserted <= 0
}

func GetArticleById(id string) (result Article, ok bool) {
	filter := bson.M{"data.content.id": id}
	err := Coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		fmt.Printf("Can't find this Article %s in DB %v \n", id, err)
		return result, false
	}
	fmt.Printf("func GetArticleById document %s exists in DB\n", result.Data.Content.Title.Short)
	return result, true
}

func DeleteAllArticles() (string, error) {
	result, err := Coll.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		return "delete all documents ERROR", err
	}
	return fmt.Sprint(result), nil
}

func UpdateOne(a Article) error {
	fmt.Println("UpdateOne start", a)
	opts := options.Update().SetUpsert(true)
	filter := bson.D{{Key: "data.content.id", Value: a.Data.Content.Id}}
	// filter := bson.D{{Key: "_id", Value: "415695b0-0dec-4d01-b8d7-e2c7ba7700ce"}}

	fmt.Println("Filter", filter)
	update := bson.D{{Key: "$set", Value: a}}
	// update := bson.D{{Key: "$set", Value: bson.D{{Key: "email", Value: "newemail@example.com"}}}}
	result, err := Coll.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}
	fmt.Printf("func Find one:\n Matched %v\n Modified %v\n Upserted %v\n UpsertedID %v\n",
		result.MatchedCount, result.ModifiedCount, result.UpsertedCount, result.UpsertedID)
	return nil
}

func (ar *ArticleRepo) BulkWrite(a []Article) error {
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

func (ar *ArticleRepo) GetNewsFromDB() []Article {
	// sort := desc
	var limit int64 = 15
	var skip int64 = 0
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(limit).SetSkip(skip)
	cursor, err := ar.coll.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		fmt.Println("func GetNewsFromDB error: ", err)
		return nil
	}
	var results []Article
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	// for _, result := range results {
	// 	fmt.Println(result)
	// }
	fmt.Println("func GetNewsFromDB ends")
	return results
}
