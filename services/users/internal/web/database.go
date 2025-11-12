package web

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	URL        = "mongodb://user_app:strong_app_password@users-service-database:27017/user_db?authSource=user_db"
	COLLECTION = "users"
	DATABASE   = "user_db"
)

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}

type Database struct {
	conn *mongo.Database
}

func NewDatabase() (*Database, error) {

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(URL))

	if err != nil {
		return nil, err
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		return nil, err
	}

	return &Database{client.Database(DATABASE)}, nil
}

func (db *Database) Register(username, password string) error {
	coll := db.conn.Collection(COLLECTION)

	var result User
	err := coll.FindOne(context.TODO(), bson.M{"username": username}, nil).Decode(&result)
	log.Println(err)

	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}

	if result.Username != "" {
		return fmt.Errorf("user already exists")
	}

	hasher := sha1.New()
	hasher.Write([]byte(password))

	user := User{
		Username: username,
		Password: hex.EncodeToString(hasher.Sum(nil)),
	}

	inserted, err := coll.InsertOne(context.TODO(), user)

	if err != nil {
		return err
	}

	log.Printf("Inserted document with _id: %v\n", inserted.InsertedID)

	return nil
}

func (db *Database) Login(username, password string) error {

	coll := db.conn.Collection(COLLECTION)

	hasher := sha1.New()
	hasher.Write([]byte(password))

	filter := bson.M{
		"username": username,
		"password": hex.EncodeToString(hasher.Sum(nil)),
	}

	log.Println(filter)

	var result User
	err := coll.FindOne(context.TODO(), filter, nil).Decode(&result)

	log.Println(err)

	if err == mongo.ErrNoDocuments {
		return fmt.Errorf("user not found")
	}

	return nil
}
