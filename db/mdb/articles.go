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

func (ar *ArticleRepo) GetByID(ctx context.Context, id string) (domain.Article, error) {
	result := domain.Article{}
	filter := bson.M{"id": id}

	err := ar.coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Printf("Can't find this Article %s in DB %v \n", id, err)

		return result, err
	}

	return result, nil
}

func (ar *ArticleRepo) Count(ctx context.Context) (int64, error) {
	docs, err := ar.coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		fmt.Println("func ArticleList CountDocuments error: ", err)

		return 0, err
	}

	return docs, nil
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

func (ar *ArticleRepo) GetArticleById(id string) (result domain.Article, ok bool) {
	filter := bson.M{"id": id}

	err := ar.coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		fmt.Printf("Can't find this Article %s in DB %v \n", id, err)

		return result, false
	}

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

	log.Printf("Mongo: inserted %v, upserted %v, deleted %v, matched %v, modified %v\n",
		res.InsertedCount, res.UpsertedCount, res.DeletedCount, res.MatchedCount, res.ModifiedCount)

	return nil
}

func (ar *ArticleRepo) GetNewsFromDB(ctx context.Context, limit, skip int) ([]domain.Article, error) {
	if limit <= 0 {
		limit = 15
	}
	// sort := date desc
	opts := options.Find().SetSort(bson.D{{Key: "dates.posted", Value: -1}}).SetLimit(int64(limit)).SetSkip(int64(skip))

	cursor, err := ar.coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Println("func GetNewsFromDB error: ", err)

		return nil, err
	}

	var results []domain.Article

	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}
