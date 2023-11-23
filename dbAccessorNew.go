package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	db  *mongo.Client
	col *mongo.Collection
}

func GetDbClient() (*MongoClient, error) {
	// TODO: SET VIA CONFIGURATION
	// - MongoUri
	// - Database Name
	//mongoURI := "mongodb://localhost:27017"
	mongoURI := "mongodb+srv://kak:ricosuave@kak-6wzzo.gcp.mongodb.net/test?retryWrites=true&w=majority"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))

	if err != nil {
		return nil, err
	}

	defer cancel()
	collection := client.Database("nebrasketball").Collection("messages")

	return &MongoClient{db: client, col: collection}, nil
}
