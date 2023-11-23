package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type GetMessagesRequest struct {
	Sender      string
	Random      bool
	FromDate    string
	ToDate      string
	MessageType string
	PageStart   int
	PageEnd     int
	PageSize    int
}

func main() {
	dbClient, err := GetDbClient()
	if err != nil {
		panic(err)
	}

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
	router.Handle("/sender/{sender}/messages", newGetMessages(dbClient)).Methods("GET")
	router.Handle("/sender/{sender}/type/{messageType}/messages", newGetMessages(dbClient)).Methods("GET")
	router.Handle("/random", randomMessage(dbClient))

	router.Handle("/randsender", randomMessageBySender(dbClient))
	router.Handle("/getallfromsender", allMessagesBySender(dbClient))
	router.Handle("/getpagedfromsender", pagedMessagesBySender(dbClient))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Google services wants 8080 or will decide for us.
		log.Printf("Defaulting to port %s", port)
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

// Possible query params:
// - FromDate (datetime)
// - ToDate (datetime)
// - random (bool)
// - pageStart
// - pageEnd
// - pageSize
func newGetMessages(s *MongoClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		queryParams := r.URL.Query()

		randomQueryParam := queryParams.Get("random")

		randomize, err := strconv.ParseBool(randomQueryParam)
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

		allMessages := dbGetMessages(s, getMessagesRequest)
	})
}
