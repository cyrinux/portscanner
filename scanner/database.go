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

func GetTaskResult(t *Task) (*ScanResult, error) {
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

	// var hostResult HostResult
	var scanResult ScanResult
	// var doc bson.M
	err = resultsCollection.FindOne(ctx, bson.M{"_id": t.Id}).Decode(&scanResult)
	if err != nil {
		log.Print(err)
	}
	log.Printf("%#v", scanResult)
	// cursor, err := resultsCollection.Find(ctx, bson.M{})
	// if err != nil {
	// 	log.Print(err)
	// }
	// for cursor.Next(context.Background()) {
	// 	// Decode the data at the current pointer and write it to data
	// 	err := cursor.Decode(&doc)
	// 	if err != nil {
	// 		log.Print(err)
	// 	}
	// 	log.Printf("%#v", doc)
	// }
	return &scanResult, err
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
