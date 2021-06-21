package scanner

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetTaskResult(t *Task) (*HostResult, error) {
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

	var hostResult *HostResult
	err = resultsCollection.FindOne(ctx, bson.M{"_id": t.Id}, options.FindOne().SetProjection(bson.M{"_id": 0})).Decode(&hostResult)
	// err = resultsCollection.FindOne(ctx, bson.M{"_id": t.Id}).Decode(&hostResult)
	if err != nil {
		log.Print(err)
	}
	return hostResult, err
}

func InsertDbResult(r *bson.M) (*mongo.InsertOneResult, error) {
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
		log.Printf("%v\n", err)
	}

	return insertOneResult, err
}
