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
	config := GetConfiguration()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(config.ConnectionString))

	defer cancel()
	collection := client.Database("nebrasketball").Collection("messages")
	messagesAccessor := &MessagesAccessor{Messages: collection}

	router := mux.NewRouter()
	handler := cors.Default().Handler(router)
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	router.Handle("/", http.FileServer(http.Dir(("./static"))))

	// Possible query params:
	// - FromDate (datetime)
	// - ToDate (datetime)
	// - random (bool)
	// - page
	// - pageSize
	// - type
	// - gifs, photos, videos, shares
	router.Handle("/messages/random", GetRandomMessage(messagesAccessor, &config)).Methods("GET")
	router.Handle("/conversations/random", GetConversation(messagesAccessor, &config)).Methods("GET")
	router.Handle("/messages/{timestamp}", GetContext(messagesAccessor, &config)).Methods("GET")
	// router.Handle("/sender/{sender}/messages", newGetMessages(dbClient)).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Google services wants 8080 or will decide for us.
	}
	log.Printf("Listening on port %s", port)

	srv := &http.Server{
		Handler:      handler,
		Addr:         fmt.Sprintf("127.0.0.1:%s", port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func GetRandomMessage(messagesAccessor *MessagesAccessor, configuration *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		participantsString := query.Get("participants")
		participants := []string{}
		if participantsString != "" {
			participants = strings.Split(participantsString, ",")
		}
		message := messagesAccessor.GetRandomMessage(participants)
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

		rawParticipants := query.Get("participants")
		if rawParticipants == "" {
			errorString := "no participants provided in query string"
			w.WriteHeader(400)
			w.Write([]byte(errorString))
		}

		participants := strings.Split(rawParticipants, ",")
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
