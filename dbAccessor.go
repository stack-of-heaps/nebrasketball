package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	KeyPassPhrase string = "fjklj4kj12414980a9fasdvklavn!@$1"
)

type GetMessagesRequest struct {
	Name      string
	Random    bool
	FromDate  primitive.DateTime
	ToDate    primitive.DateTime
	PageStart int
	PageEnd   int
	PageSize  int
}

type Message struct {
	Sender    string
	Timestamp int
	Content   string
	Photos    []Photo
	Reactions []Reaction
	Share     Share
	Type      string
}

type Participants struct {
	Name string
}

type Reaction struct {
	Reaction string
	Actor    string
}

type Photo struct {
	Uri      string
	Creation int
}

type Share struct {
	Link string
}

type Messages struct {
	Messages []Message `json:"messages"`
}

func pagedPipelineBuilder(sender string, startingId string, limit int) []bson.M {

	//pipeline := []bson.M{bson.M{"$match": bson.M{"sender": sender, "_id": bson.M{"$gt": startingId}}}, bson.M{"$limit": maxItems}}
	matchElement := matchPipelineBuilder(sender, startingId)
	limitElement := bson.M{"$limit": limit}
	pipeline := []bson.M{matchElement, limitElement}

	return pipeline
}

func pagedMessagesLogic(s *MongoClient, sender string, startingId string) (messages Messages, encryptedLastId string) {

	fmt.Println("Begin pagedMessagesBySender()")

	// pipeline := []bson.D{bson.D{{"$match", {bson.D{{"sender", sender}}}}}, bson.D{{"$limit", maxItems}}}
	// pipeline := []bson.M{bson.M{"$match": bson.M{"sender": sender, "_id": bson.M{"$gt": startingId}}}, bson.M{"$limit": maxItems}}

	maxItems := 10
	pipeline := pagedPipelineBuilder(sender, startingId, maxItems)
	cursor, _ := s.col.Aggregate(context.Background(), pipeline)

	var messageBatch Messages
	var result Message
	var rawId bson.RawValue

	for cursor.Next(context.Background()) {
		cursorErr := cursor.Decode(&result)
		if cursorErr != nil {
			log.Fatal("Error in pagedMessagesBySender() cursor: ", cursorErr)
		}

		messageBatch.Messages = append(messageBatch.Messages, result)
		rawId = cursor.Current.Lookup("_id")
	}

	lastId := stringFromRawValue(rawId)

	encryptedLastId = encryptLastId(lastId)

	return messageBatch, encryptedLastId
}

func encryptLastId(lastId string) string {

	fmt.Println("Beginning encryptLastId()")

	// Generate AES cipher with 32 byte passphrase
	aesCipher, err := aes.NewCipher([]byte(KeyPassPhrase))

	if err != nil {
		fmt.Println("Error in encryptLastId(): ", err)
	}

	// GCM "Galois/Counter Mode": Symmetric Keyy cryptographic block cipher
	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		fmt.Println("Error in encryptLastId(): ", err)
	}

	// Nonce is literally a "one off" byte array which will be populated by a random sequence below.
	// The nonce is prepended/appended to the cipher (?) and is used in deciphering
	nonce := make([]byte, gcm.NonceSize())

	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println("Error in io.ReadFull: ", err)
	}

	encryptedByteArray := gcm.Seal(nonce, nonce, []byte(lastId), nil)

	// Convert to Base64 to ensure we can transmit via HTTP without error or corruption
	encryptedString := base64.StdEncoding.EncodeToString(encryptedByteArray)

	fmt.Println("Ending encryptLastId()")

	return encryptedString
}

func stringFromRawValue(rawId bson.RawValue) string {
	objectID := rawId.ObjectID().String()
	lastId := strings.Split(objectID, "\"")

	return lastId[1]
}

func matchPipelineBuilder(sender string, startingId string) bson.M {

	matchRoot := bson.M{"$match": ""}
	senderElement := bson.M{"sender": sender}
	idElement := bson.M{"_id": ""}
	gtElement := bson.M{"$gt": ""}

	if startingId == "" {
		matchRoot["$match"] = senderElement
	} else {
		gtElement["$gt"] = startingId
		idElement["_id"] = gtElement
		tempArray := []bson.M{senderElement, idElement}
		matchRoot["$match"] = tempArray
	}

	fmt.Println("Returning matchroot: ", matchRoot)
	return matchRoot
}

func getPipeline(category string, queryToMatch string, isRandom bool) []primitive.D {

	if isRandom {
		if category == "" {
			return []bson.D{{{"$sample", bson.D{{"size", 1}}}}}
		} else {
			return []bson.D{
				{{"$match", bson.D{{category, queryToMatch}}}},
				{{"$sort", bson.D{{"timestamp", -1}}}}}
		}
	}

	return []bson.D{
		bson.D{{"$match", bson.D{{category, queryToMatch}}}},
		bson.D{{"$sample", bson.D{{"size", 1}}}}}
}

func dbGetMessages(
	mongoClient *MongoClient,
	getMessageRequest GetMessagesRequest) []Mesage {

	pipeline := getPipeline(category, queryToMatch, isRandom)
	cursor, _ := mongoClient.col.Aggregate(context.Background(), pipeline)

	var allMessages Messages

	for cursor.Next(context.Background()) {
		var result Message
		cursorErr := cursor.Decode(&result)
		if cursorErr != nil {
			panic("Error in cursor")
		}

		allMessages.Messages = append(allMessages.Messages, result)
	}

	return allMessages.Messages
}

func checkForVideo(obj Message) bool {

	if obj.Photos == nil {
		return false
	}

	path := obj.Photos[0].Uri
	ext := ".mp4"
	fmt.Println("Path: ", path)
	return strings.Contains(path, ext)
}

func handleMediaPath(origPhotos []Photo) []Photo {

	if origPhotos == nil || origPhotos[0].Uri == "" {
		return origPhotos
	}

	path := &origPhotos[0].Uri
	videos := "/videos/"
	photos := "/photos/"
	gifs := "/gifs/"

	if strings.Contains(*path, videos) {
		*path = stripPath(*path, videos)
	}

	if strings.Contains(*path, photos) {
		*path = stripPath(*path, photos)
	}

	if strings.Contains(*path, gifs) {
		*path = stripPath(*path, gifs)
	}

	return origPhotos
}

func stripPath(path string, substringToStrip string) string {
	pathIndex := strings.Index(path, substringToStrip)
	return path[pathIndex:]
}
