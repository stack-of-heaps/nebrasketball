package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type GetMessagesRequest struct {
	Sender   string
	Type     string
	FromDate string
	ToDate   string
	Page     int
	PageSize int
}

func main() {

	fmt.Println("starting main")
	config := GetConfiguration()
	fmt.Println("got config: ", config)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.Db.ConnectionString))
	if err != nil {
		fmt.Println("mongo connect err: ", err)
	}
	defer cancel()

	collection := client.Database(config.Db.Database).Collection(config.Db.Collection)
	messagesAccessor := &MessagesAccessor{Messages: collection}

	router := mux.NewRouter()
	handler := cors.Default().Handler(router)
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	router.Handle("/", http.FileServer(http.Dir(("./static"))))

	// Query params:
	// - random (bool)
	// - participants (comma-separated string)
	// TODO
	// - FromDate (int64: ms from epoch)
	// - ToDate (int64: ms from epoch)
	// - gifs, photos, videos, shares
	router.Handle("/messages/random", GetRandomMessage(messagesAccessor, &config)).Methods("GET")
	router.Handle("/conversations/random", GetConversation(messagesAccessor, &config)).Methods("GET")
	router.Handle("/messages/{timestamp}", GetContext(messagesAccessor, &config)).Methods("GET")
	// router.Handle("/sender/{sender}/messages", newGetMessages(dbClient)).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Handler:      handler,
		Addr:         fmt.Sprintf("%s:%s", config.ServerAddress, port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
	log.Printf("Listening on port %s", port)
}

func GetRandomMessage(messagesAccessor *MessagesAccessor, configuration *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		fmt.Println("query: ", query)
		participantsString := query.Get("participants")
		participants := []string{}
		if participantsString != "" {
			participants = getParticipants(participantsString, configuration)
		}

		filtersString := query.Get("filters")
		filters := []string{}
		if filtersString != "" {
			filters = strings.Split(filtersString, ",")
		}

		fmt.Println("query filters: ", filtersString)
		fmt.Println("filters: ", filters)

		message := messagesAccessor.GetRandomMessage(participants, filters)
		json, _ := json.Marshal(message)

		w.Write(json)
	}
}

func GetConversation(messagesAccessor *MessagesAccessor, configuration *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		fuzzFactor := 0

		fuzzFactorStr := query.Get("fuzzFactor")
		if fuzzFactorStr != "" {
			parsedFuzzFactor, err := strconv.ParseInt(fuzzFactorStr, 10, 64)

			if err != nil {
				fmt.Printf("cannot convert '%s' to int", fuzzFactorStr)
			}

			fuzzFactor = int(parsedFuzzFactor)
		}

		participantsString := query.Get("participants")
		participants := []string{}
		if participantsString != "" {
			participants = getParticipants(participantsString, configuration)
		}

		messages := messagesAccessor.GetConversation(participants, fuzzFactor)
		json, _ := json.Marshal(messages)

		w.Write(json)
	}
}

func GetContext(messagesAccessor *MessagesAccessor, configuration *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timestamp := mux.Vars(r)["timestamp"]
		if timestamp == "nil" {
			fmt.Printf("Timestamp received: %v", timestamp)
			w.WriteHeader(400)
		}

		timestampAsInt, err := strconv.ParseInt(timestamp, 10, 64)

		if err != nil {
			w.WriteHeader(400)
		}

		context := messagesAccessor.GetContext(timestampAsInt)
		json, _ := json.Marshal(context)
		w.Write(json)
	}
}

// Possible query params:
// - FromDate (datetime)
// - ToDate (datetime)
// - random (bool)
// - pageStart
// - pageEnd
// - pageSize
func GetMessages(messagesAccessor *MessagesAccessor, configuration *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		queryParams := r.URL.Query()

		pageSize, err := strconv.Atoi(queryParams.Get("pageSize"))
		if err != nil {
			pageSize = 0
		}

		page, err := strconv.Atoi(queryParams.Get("page"))
		if err != nil {
			page = 0
		}

		getMessagesRequest := GetMessagesRequest{
			Sender:   queryParams.Get("sender"),
			FromDate: queryParams.Get("fromDate"),
			ToDate:   queryParams.Get("toDate"),
			PageSize: pageSize,
			Page:     page,
			Type:     queryParams.Get("type"),
		}

		// TEMP
		messages := messagesAccessor.GetMessages((getMessagesRequest))
		messagesJson, err := json.Marshal(messages)

		w.Write(messagesJson)
	}
}

func getParticipants(participantsQuery string, config *Configuration) []string {
	unmappedParticipants := strings.Split(participantsQuery, ",")
	participants := []string{}
	for _, p := range unmappedParticipants {
		participants = append(participants, config.Participants[p])
	}
	return participants
}
