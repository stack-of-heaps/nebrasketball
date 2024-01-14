package main

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type M = bson.M

type MessagesAccessor struct {
	Messages *mongo.Collection
}

func (messagesAccessor *MessagesAccessor) GetMessages(getMessagesRequest GetMessagesRequest) []Message {
	pipeline := []bson.D{}

	if getMessagesRequest.Sender != "" {
		pipeline = append(pipeline, getPipelineElement("name", getMessagesRequest.Sender))
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

func (messagesAccessor *MessagesAccessor) GetRandomMessage() Message {
	pipeline := []M{{"$sample": M{"size": 1}}}
	cursor, _ := messagesAccessor.Messages.Aggregate(context.Background(), pipeline)
	cursor.Next(context.Background())
	message := Message{}
	cursor.Decode(&message)

	return message
}

func (messagesAccessor *MessagesAccessor) GetConversation(participants []string, fuzzFactor int) []Message {
	pipeline := []M{
		{
			"$match": M{
				"sender_name": M{"$in": participants},
			},
		},
		{"$sample": M{"size": 1}},
	}

	cursor, err := messagesAccessor.Messages.Aggregate(context.Background(), pipeline)
	if err != nil {
		fmt.Print("Mongo error in GET CONVERSATIONS: ", err)
	}

	startingMessage := Message{}
	cursor.Next(context.Background())
	cursor.Decode(&startingMessage)

	MaxAttempts := 5
	counter := 0
	messages := []Message{}

	for counter <= MaxAttempts {
		counter += 1
		fmt.Println("COUNTER: ", counter)

		var wg sync.WaitGroup
		wg.Add(2)

		go messagesAccessor.goForwards(
			&wg,
			startingMessage.Timestamp,
			int64(counter)*10*60,
			&participants,
			&messages,
			fuzzFactor)

		go messagesAccessor.goBackwards(
			&wg,
			startingMessage.Timestamp,
			int64(counter)*10*60,
			&participants,
			&messages,
			fuzzFactor)

		wg.Wait()

		if allParticipantsFound(&participants, &messages) {
			slices.SortFunc(messages, compareTimestampValue)

			return messages
		}
	}

	fmt.Println("No conversation found. Staritng GetConversation over again.")
	return messagesAccessor.GetConversation(participants, fuzzFactor)
}

func (messagesAccessor *MessagesAccessor) goBackwards(
	wg *sync.WaitGroup,
	startTime int64,
	seconds int64,
	participants *[]string,
	allMessages *[]Message,
	fuzzFactor int) {
	greaterThanTimestamp := startTime - (seconds * 1000)
	if greaterThanTimestamp < 0 {
		greaterThanTimestamp = 0
	}

	defer wg.Done()

	pipeline := []M{
		{
			"$match": M{
				"timestamp_ms": M{"$gt": greaterThanTimestamp, "$lt": startTime},
			},
		},
		{
			"$sort": M{"timestamp_ms": -1},
		},
		{
			"$limit": 100,
		},
	}

	cursor, err := messagesAccessor.Messages.Aggregate(context.Background(), pipeline)
	if err != nil {
		fmt.Println("Error in gobackwards: ", err)
	}

	shouldContinue, greaterThanTimestamp := getConversationMessages(cursor, participants, allMessages, greaterThanTimestamp, fuzzFactor)

	if shouldContinue {
		wg.Add(1)
		messagesAccessor.goBackwards(wg, greaterThanTimestamp, 5*60*1000, participants, allMessages, fuzzFactor)
	}
}

func (messagesAccessor *MessagesAccessor) goForwards(
	wg *sync.WaitGroup,
	startTime int64,
	seconds int64,
	participants *[]string,
	allMessages *[]Message,
	fuzzFactor int) {

	defer wg.Done()

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

	shouldContinue, lessThanTimestamp := getConversationMessages(cursor, participants, allMessages, lessThanTimestamp, fuzzFactor)

	if shouldContinue {
		wg.Add(1)
		messagesAccessor.goForwards(wg, lessThanTimestamp, 5*60*1000, participants, allMessages, fuzzFactor)
	}
}

func getConversationMessages(
	cursor *mongo.Cursor,
	participants *[]string,
	allMessages *[]Message,
	timestamp int64,
	fuzzFactor int) (bool, int64) {

	const MaxFuzz = 5
	fuzzCount := 0
	lastChance := false
	for cursor.Next(context.Background()) {
		currentMessage := Message{}
		cursor.Decode(&currentMessage)
		if !slices.Contains(*participants, currentMessage.Sender) {
			if fuzzFactor == 0 || lastChance {
				return false, currentMessage.Timestamp
			} else if fuzzCount < MaxFuzz {
				fuzzCount++
				if fuzzCount >= MaxFuzz {
					lastChance = true
				}
			}
		}

		*allMessages = append(*allMessages, currentMessage)
	}

	return true, timestamp
}

func allParticipantsFound(participants *[]string, messages *[]Message) bool {
	foundParticipants := []string{}
	for _, value := range *messages {

		if slices.Contains(*participants, value.Sender) && !slices.Contains(foundParticipants, value.Sender) {
			foundParticipants = append(foundParticipants, value.Sender)
		}

		if len(foundParticipants) == len(*participants) {
			return true
		}
	}

	return false
}

func compareTimestampValue(a, b Message) int {
	if a.Timestamp < b.Timestamp {
		return -1
	}

	if a.Timestamp == b.Timestamp {
		return 0
	}

	return 1
}

func getPipelineElement(elementKey string, elementValue string) bson.D {
	return bson.D{{"$match", bson.D{{elementKey, elementValue}}}}
}
