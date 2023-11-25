package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	db  *mongo.Client
	col *mongo.Collection
}

// TODO:
//   - Consider generating a dbclient only after successfully validated message
//   - Once I have the database set up again locally and can test this basic implementation,
//     Then change method so that it accepts GetMessageRequest and can handle the request in a more
//     sophisticated fashion.
func GetMessage(dbClient *MongoClient, senderName string, encryptedId string) Message {

	message := Message{}

	if encryptedId != "" {
		decryptedId := decryptId(encryptedId)
		dbClient.col.FindOne(context.Background(), bson.D{{"_id", decryptedId}}).Decode(&message)

		return message
	}

	return GetRandomMessage(dbClient, senderName)
}

func GetRandomMessage(dbClient *MongoClient, senderName string) Message {

	pipeline := []bson.D{
		bson.D{{"$match", bson.D{{"sender", senderName}}}},
		bson.D{{"$sample", bson.D{{"size", 1}}}}}

	cursor, _ := dbClient.col.Aggregate(context.Background(), pipeline)
	cursor.Next(context.Background())
	message := Message{}
	cursor.Decode(&message)

	return message
}

func decryptId(encryptedId string) string {
	c, err := aes.NewCipher([]byte(KeyPassPhrase))
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		panic(err)
	}

	nonceSize := gcm.NonceSize()

	nonce, encryptedId := encryptedId[:nonceSize], encryptedId[nonceSize:]
	decryptedId, err := gcm.Open(nil, []byte(nonce), []byte(encryptedId), nil)
	if err != nil {
		panic(err)
	}

	return string(decryptedId)
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
