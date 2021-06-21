package scanner

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func GetTaskResult(t *Task) ([]byte, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://admin:secret@localhost:27017/?authSource=admin"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	defer cancel()

	scannerDatabase := client.Database("scanner")
	resultsCollection := scannerDatabase.Collection("results")

	mongoRow := bson.D{primitive.E{
		Key: t.Id,
	}}

	cursor, err := resultsCollection.Find(ctx, mongoRow)
	if err != nil {
		log.Fatal(err)
	}

	var results bson.D

	if err = cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}
	log.Println(results)

	resultBytes, err := bson.Marshal(results)

	return resultBytes, err

}

func InsertDbResult(r *bson.D) (*mongo.InsertOneResult, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://admin:secret@localhost:27017/?authSource=admin"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	defer cancel()

	scannerDatabase := client.Database("scanner")
	resultsCollection := scannerDatabase.Collection("results")

	insertOneResult, err := resultsCollection.InsertOne(ctx, r)
	if err != nil {
		log.Printf("%v\n", insertOneResult)
	}

	return insertOneResult, err
}
