package internal

import (
	"context"
	"eda-logs/internal/types"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	URL        = "mongodb://user_app:strong_app_password@logs-service-database:27018/user_db?authSource=user_db"
	COLLECTION = "logs"
	DATABASE   = "log_db"
)

type Log struct {
	Message string `bson:"message"`
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

func (db *Database) Save(data types.Log) error {
	coll := db.conn.Collection(COLLECTION)

	inserted, err := coll.InsertOne(context.TODO(), data)

	if err != nil {
		return err
	}

	log.Printf("Inserted document with _id: %v\n", inserted.InsertedID)

	return nil
}
