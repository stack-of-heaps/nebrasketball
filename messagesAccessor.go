package main

import (
	"context"
	"fmt"
	"slices"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type M = bson.M

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
	// Adapt window function
	// use this link
	// https://stackoverflow.com/questions/12012253/finding-the-next-document-in-mongodb
	pipeline := []M{
		{
			"$match": M{
				"sender_name": M{
					"$in": participants,
				},
			},
		},
		M{"$sample": M{"size": 1}},
	}

	cursor, err := messagesAccessor.Messages.Aggregate(context.Background(), pipeline)
	if err != nil {
		fmt.Print("Mongo error in GET CONVERSATIONS: ", err)
	}

	message := Message{}
	cursor.Next(context.Background())
	cursor.Decode(&message)

	messages := []Message{}
	MaxAttempts := 5
	counter := 0
	shouldGoForwards := true
	shouldGoBackwards := true

	goForwardsTimestamp := message.Timestamp
	goBackwardsTimestamp := message.Timestamp
	for counter <= MaxAttempts {
		counter += 1
		fmt.Println("COUNTER: ", counter)

		for shouldGoForwards {
			shouldGoForwards, goForwardsTimestamp = messagesAccessor.GoForwards(goForwardsTimestamp, int64(counter)*10*60, &participants, &messages)
		}

		for shouldGoBackwards {
			shouldGoBackwards, goBackwardsTimestamp = messagesAccessor.GoBackwards(goBackwardsTimestamp, int64(counter)*10*60, &participants, &messages)
		}

		if AllParticipantsFound(&participants, &messages) {
			slices.SortFunc(messages, CompareTimestampValue)

			return messages
		}
	}

	return messagesAccessor.GetConversation(participants)
}

func AllParticipantsFound(participants *[]string, messages *[]Message) bool {
	foundParticipants := []string{}
	for _, value := range *messages {

		if !slices.Contains(foundParticipants, value.Sender) {
			foundParticipants = append(foundParticipants, value.Sender)
		}

		if len(foundParticipants) == len(*participants) {
			return true
		}
	}

	return false
}

func (messagesAccessor *MessagesAccessor) GoBackwards(startTime int64, seconds int64, participants *[]string, allMessages *[]Message) (bool, int64) {
	greaterThanTimestamp := startTime - (seconds * 1000)
	if greaterThanTimestamp < 0 {
		greaterThanTimestamp = 0
	}

	pipeline := []M{
		{
			"$match": M{
				"timestamp_ms": M{"$gt": greaterThanTimestamp, "$lt": startTime},
			},
		},
		{"$sort": M{"timestamp_ms": -1}},
		{"$limit": 100},
	}

	cursor, err := messagesAccessor.Messages.Aggregate(context.Background(), pipeline)
	if err != nil {
		fmt.Println("Error in gobackwards: ", err)
	}

	return GetConversationMessages(cursor, participants, allMessages, greaterThanTimestamp)
}

func (messagesAccessor *MessagesAccessor) GoForwards(startTime int64, seconds int64, participants *[]string, allMessages *[]Message) (bool, int64) {
	lessThanTimestamp := startTime + (seconds * 1000)
	pipeline := []M{
		{
			"$match": M{
				"timestamp_ms": M{"$gt": startTime, "$lt": lessThanTimestamp},
			},
		},
		{
			"$sort": M{"timestamp_ms": 1},
		},
		{
			"$limit": 100,
		},
	}

	cursor, err := messagesAccessor.Messages.Aggregate(context.Background(), pipeline)
	if err != nil {
		fmt.Println("error in go forwards: ", err)
	}

	return GetConversationMessages(cursor, participants, allMessages, lessThanTimestamp)
}

func GetConversationMessages(cursor *mongo.Cursor, participants *[]string, allMessages *[]Message, timestamp int64) (bool, int64) {
	currentMessage := Message{}
	for cursor.Next(context.Background()) {
		cursor.Decode(&currentMessage)
		if !slices.Contains(*participants, currentMessage.Sender) {
			return false, timestamp
		}

		*allMessages = append(*allMessages, currentMessage)
	}

	return true, timestamp
}

func GetAllMessages(cursor *mongo.Cursor, err error, loggingMessage string) []Message {
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			fmt.Println("No documents found ", loggingMessage)
			return []Message{}
		default:
			fmt.Println(loggingMessage, err)
		}
	}

	messages := []Message{}
	for cursor.Next(context.Background()) {
		message := Message{}
		cursor.Next(context.Background())
		cursor.Decode(&message)
		messages = append(messages, message)
	}

	return messages
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

func CompareTimestampValue(a, b Message) int {
	if a.Timestamp > b.Timestamp {
		return 1
	}

	if a.Timestamp == b.Timestamp {
		return 0
	}

	return 0
}

func GetPipelineElement(elementKey string, elementValue string) bson.D {
	return bson.D{{"$match", bson.D{{elementKey, elementValue}}}}
}
