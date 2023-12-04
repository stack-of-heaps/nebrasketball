package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
)

type GetMessagesRequest struct {
	Name        string
	Random      bool
	FromDate    string
	ToDate      string
	MessageType string
	PageStart   int
	PageEnd     int
	PageSize    int
}

func main() {
	mongoURI := "mongodb+srv://kak:ricosuave@kak-6wzzo.gcp.mongodb.net/test?retryWrites=true&w=majority"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))

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
	// - pageStart
	// - pageEnd
	// - number
	router.Handle("/messages/random", GetRandomMessage(messagesAccessor)).Methods("GET")
	// router.Handle("/sender/{sender}/messages", newGetMessages(dbClient)).Methods("GET")
	// router.Handle("/sender/{sender}/type/{messageType}/messages", newGetMessages(dbClient)).Methods("GET")

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

		randomize, err := strconv.ParseBool(queryParams.Get("random"))
		if err != nil {
			randomize = false
		}

		pageSize, err := strconv.ParseInt(queryParams.Get("pageSize"), 10, 16)
		if err != nil {
			pageSize = 0
		}

		pageStart, err := strconv.ParseInt(queryParams.Get("pageStart"), 10, 16)
		if err != nil {
			pageStart = 0
		}

		pageEnd, err := strconv.ParseInt(queryParams.Get("pageEnd"), 10, 16)
		if err != nil {
			pageEnd = 0
		}

		getMessagesRequest := GetMessagesRequest{
			Name:      queryParams.Get("name"),
			Random:    randomize,
			FromDate:  queryParams.Get("fromDate"),
			ToDate:    queryParams.Get("toDate"),
			PageSize:  int(pageSize),
			PageStart: int(pageStart),
			PageEnd:   int(pageEnd),
		}

		// TEMP
		messages := messagesAccessor.GetMessages((getMessagesRequest))
		messagesJson, err := json.Marshal(messages)

		w.Write(messagesJson)
	}
}
