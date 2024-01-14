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
	// mongoURI := "mongodb+srv://kak:ricosuave@kak-6wzzo.gcp.mongodb.net/test?retryWrites=true&w=majority"
	mongoUri := "mongodb://localhost:27017"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))

	defer cancel()
	collection := client.Database("nebrasketball").Collection("messages")
	messagesAccessor := &MessagesAccessor{Messages: collection}

	router := mux.NewRouter()
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
	router.Handle("/messages/random", GetRandomMessage(messagesAccessor)).Methods("GET")
	router.Handle("/conversations/random", GetConversation(messagesAccessor)).Methods("GET")
	// router.Handle("/sender/{sender}/messages", newGetMessages(dbClient)).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Google services wants 8080 or will decide for us.
	}
	log.Printf("Listening on port %s", port)

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("127.0.0.1:%s", port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func GetRandomMessage(messagesAccessor *MessagesAccessor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		message := messagesAccessor.GetRandomMessage()
		json, _ := json.Marshal(message)

		w.Write(json)
	}
}

func GetConversation(messagesAccessor *MessagesAccessor) http.HandlerFunc {
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

// Possible query params:
// - FromDate (datetime)
// - ToDate (datetime)
// - random (bool)
// - pageStart
// - pageEnd
// - pageSize
func GetMessages(messagesAccessor *MessagesAccessor) http.HandlerFunc {
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
