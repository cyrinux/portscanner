package scanner

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoTask struct {
	ID     primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Result AllScanResults     `bson:"result" json:"result"`
}

func GetTaskResult(t *Task) (*AllScanResults, error) {
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

	var result bson.M
	// var allScanResults AllScanResults
	var mongoTask *MongoTask
	err = resultsCollection.FindOne(ctx, bson.M{"_id": t.Id}).Decode(&result)
	// err = resultsCollection.FindOne(ctx, bson.M{"_id": t.Id}).Decode(&allScanResults)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Result 1: %v\n", result)

	// for cursor.Next(ctx) {
	// 	var results bson.M
	// 	if err = cursor.Decode(&results); err != nil {
	// 		log.Fatal(err)
	// 	}
	// log.Printf("%+v", result)
	// }

	bsonBytes, err := bson.Marshal(&result)
	if err != nil {
		log.Fatal(err)
	}
	err = bson.Unmarshal(bsonBytes, &mongoTask)
	if err != nil {
		log.Fatal(err)
	}

	// cursor.Close(ctx)
	log.Printf("MongoTask 1: %+v\n", mongoTask)
	log.Printf("MongoTask 2: %+v\n", mongoTask.Result.GetHostResult())
	// var allScanResults &AllScanResults
	return &mongoTask.Result, err
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
		log.Printf("%v\n", insertOneResult)
	}
	log.Printf("%v\n", insertOneResult)

	return insertOneResult, err
}
