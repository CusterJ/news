package mdb

import (
	"News/domain"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepo struct {
	coll *mongo.Collection
	User domain.User
}

func NewUserRepo(coll *mongo.Collection) *UserRepo {
	return &UserRepo{
		coll: coll,
		User: domain.User{},
	}
}

func (ur *UserRepo) UserSave(u domain.User) error {
	fmt.Println("func UserSave -> start")

	_, err := ur.coll.InsertOne(context.TODO(), u)
	if err != nil {
		return fmt.Errorf("func UserSave -> mongo InserOne error: %v", err)
	}

	return nil
}

func (ur *UserRepo) UserExistsInDB(username string) (domain.User, bool) {
	fmt.Println("func UserExistsInDB -> start")

	filter := bson.M{"name": username}
	err := ur.coll.FindOne(context.TODO(), filter).Decode(&ur.User)
	if err != nil {
		fmt.Printf("func UserExistsInDB -> Can't find this User %s in DB %v \n", username, err)
		return ur.User, false
	}

	fmt.Printf("func UserExistsInDB -> document %+v exists in DB\n", ur.User)
	return ur.User, true
}

func (ur *UserRepo) UserFind(username, password string) error {
	fmt.Println("func UserFind -> start")

	user, exists := ur.UserExistsInDB(username)
	if exists {
		return fmt.Errorf("user already exists in db")
	}

	fmt.Println(user)

	return nil
}

func (ur *UserRepo) UserUpdate() {

}

func (ur *UserRepo) UserDelete() {

}
