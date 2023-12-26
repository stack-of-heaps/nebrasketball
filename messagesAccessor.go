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

	const MaxAttempts = 5
	foundParticipants := []string{}
	messages := []Message{}
	message := Message{}
	cursor.Next(context.Background())
	cursor.Decode(&message)
	counter := 0
	foundParticipants = append(foundParticipants, message.Sender)
	fmt.Println("BEFORE LOOP FOUNDPARTICIPANTS: ", foundParticipants)
	for len(foundParticipants) != len(participants) {
		fmt.Println("COUNTER VALUE: ", counter)
		counter += 1

		if counter == MaxAttempts {
			break
		}

		backwardsMessages := messagesAccessor.GoBackwards(message.Timestamp, int64(counter)*5*60)
		fmt.Println("BACKWARDS MESSAGES LENGTH: ", len(backwardsMessages))
		for _, message := range backwardsMessages {
			if slices.Contains(participants, message.Sender) && !slices.Contains(foundParticipants, message.Sender) {
				foundParticipants = append(foundParticipants, message.Sender)
			}

			messages = append(messages, backwardsMessages...)
		}

		forwardsMessages := messagesAccessor.GoForwards(message.Timestamp, int64(counter)*5*60)
		fmt.Println("FORWARDS MESSAGES LENGTH: ", len(forwardsMessages))
		for _, message := range forwardsMessages {
			if slices.Contains(participants, message.Sender) && !slices.Contains(foundParticipants, message.Sender) {
				foundParticipants = append(foundParticipants, message.Sender)
			}

			messages = append(messages, forwardsMessages...)
		}

		fmt.Println("FOUNDPARTICIPANTS END OF LOOP: ", counter, foundParticipants)
		fmt.Println("LENGTH FOUNDPARTICIPANTS != LENGTH PARTICIPANTS: ", len(participants) != len(foundParticipants))
	}

	fmt.Println("FOUND PARTICIPANTS: ", foundParticipants)
	// fmt.Print(messages)

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

func (messagesAccessor *MessagesAccessor) GoBackwards(startTime int64, seconds int64) []Message {
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
		{
			"$sort": M{"timestamp_ms": -1},
		},
	}

	cursor, err := messagesAccessor.Messages.Aggregate(context.Background(), pipeline)
	return GetAllMessages(cursor, err, "In GoBackwards()")
}

func (messagesAccessor *MessagesAccessor) GoForwards(startTime int64, seconds int64) []Message {
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
	}

	cursor, err := messagesAccessor.Messages.Aggregate(context.Background(), pipeline)
	return GetAllMessages(cursor, err, "In GoForwards()")
}

func GetAllMessages(cursor *mongo.Cursor, err error, loggingMessage string) []Message {
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			fmt.Println("No documents found going backwards!")
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

func GetPipelineElement(elementKey string, elementValue string) bson.D {
	return bson.D{{"$match", bson.D{{elementKey, elementValue}}}}
}
