package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessagesAccessor struct {
	Messages *mongo.Collection
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

	if getMessagesRequest.Name != "" {
		pipeline = append(pipeline, GetPipelineElement("name", getMessagesRequest.Name))

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
