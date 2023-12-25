package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Participants struct {
	Name string
}

type Reaction struct {
	Reaction string
	Actor    string
}

type Photo struct {
	Uri      string
	Creation int64 `bson:"timestamp_ms"`
}

type Video struct {
	Uri      string
	Creation int64 `bson:"timestamp_ms"`
}

type Gif struct {
	Uri      string
	Creation int64 `bson:"timestamp_ms"`
}

type Share struct {
	Link      string
	Sharetext string
}

type Messages struct {
	Messages []Message
}

type Message struct {
	Id        primitive.ObjectID `bson:"_id"`
	Sender    string             `bson:"sender_name"`
	Timestamp int64              `bson:"timestamp_ms"`
	Content   string
	Photos    []Photo
	Reactions []Reaction
	Gifs      []Gif
	Videos    []Video
	Share     Share
	Type      string
}

type MessagesAccessor struct {
	Messages *mongo.Collection
}

func (messagesAccessor *MessagesAccessor) GetConversation(participants []string) []Message {
	messages := []Message{}
	// cursor, _ := messagesAccessor.Messages.Find(context.Background(), bson.M{"sender_name": bson.M{"$in": participants}})

	// Adapt window function
	// use this link
	// https://stackoverflow.com/questions/12012253/finding-the-next-document-in-mongodb
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"sender_name": bson.M{
					"$in": participants,
				},
			},
		},
		{
			"$setWindowFields": bson.M{
				"sortBy": bson.M{"_id": 1},
				"output": bson.M{
					"next": bson.M{
						"$push": "$$ROOT", "window": bson.M{"documents": []int{1, 1}},
					},
				},
			},
		},
	}

	// pipeline_ := []bson.M{"$setWindowFields": bson.M{"partitionBy": "timestamp_ms", "sortBy": bson.M{"timestamp_ms": 1}}, "output": bson.M{"previous": bson.M{"$push": "$$ROOT"}}}
	cursor, err := messagesAccessor.Messages.Aggregate(context.Background(), pipeline)
	if err != nil {
		fmt.Print("Mongo error: ", err)
	}

	cursor.Next(context.Background())

	/*
		for i := 0; i < 5; i++ {
			message := Message{}
			cursor.Next(context.Background())
			cursor.Decode(&message)
			messages = append(messages, message)
		}
	*/
	message := Message{}

	poop, _ := cursor.Current.Elements()
	fmt.Print("ELEMENTS: ", poop)

	cursor.Decode(&message)
	messages = append(messages, message)

	fmt.Print(messages)

	return messages

	// Has to know how to get a grouping of messages
	// Sender names (at least 2) OR timeframe
	// PSEUDOCODE
	// Start with arbitrary timeframe
	// Find one name at random
	// Expand time frame by 5 minutes
	// If match, consider returning
	// If additional name match, but not yet all, expand by 5 mins
	// If no match, expand by 60 mins
	// If all match, contract by 15 mins
	// If all match, contract by 15 mins
	// ... Just an example of how this could be done.
	// Need to consider when you would scrap initial random message search and start with a new random one
}

func (messagesAccessor *MessagesAccessor) GetRandomMessage() Message {

	pipeline := []bson.D{bson.D{{"$sample", bson.D{{"size", 1}}}}}
	cursor, _ := messagesAccessor.Messages.Aggregate(context.Background(), pipeline)
	cursor.Next(context.Background())
	message := Message{}
	cursor.Decode(&message)

	return message
}

func (messagesAccessor *MessagesAccessor) GetMessages(getMessagesRequest GetMessagesRequest) []Message {
	pipeline := []bson.D{}

	if getMessagesRequest.Sender != "" {
		pipeline = append(pipeline, GetPipelineElement("name", getMessagesRequest.Sender))
	}

	var messages []Message
	cursor, _ := messagesAccessor.Messages.Aggregate(context.Background(), pipeline)

	for cursor.Next(context.Background()) {
		var result Message
		cursorErr := cursor.Decode(&result)
		if cursorErr != nil {
			panic("Error in cursor")
		}

		messages = append(messages, result)
	}

	return messages
}

func GetPipelineElement(elementKey string, elementValue string) bson.D {
	return bson.D{{"$match", bson.D{{elementKey, elementValue}}}}
}
