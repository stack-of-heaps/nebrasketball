package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	db  *mongo.Client
	col *mongo.Collection
}
type Participants struct {
	Name string
}

type Reaction struct {
	Reaction string
	Actor    string
}

type Photo struct {
	Uri      string
	Creation int
}

type Share struct {
	Link string
}

type Messages struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Sender    string
	Timestamp int
	Content   string
	Photos    []Photo
	Reactions []Reaction
	Share     Share
	Type      string
}

type ServerResponse struct {
	MessageResults Messages
	Empty          bool
	FirstID        primitive.ObjectID
	LastID         primitive.ObjectID
}

// Only grabs alphanumeric ID and quotes between ObjectID()
var ObjectIdRegEx = regexp.MustCompile(`"(.*?)"`)

func reformatObjectId(objectId string) string {

	fmt.Println("objectID passed in: ", objectId)
	var idStringBeginning = "ObjectId("
	var idStringEnd = ")"
	id := ObjectIdRegEx.FindString(objectId)

	if id == "" {
		fmt.Println("Error in reformatObjectId")
		return ""
	}

	return idStringBeginning + id + idStringEnd
}

func randomMessage(s *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Construct aggregation "pipeline" to return 1 random document from entire collection
		pipeline := []bson.D{bson.D{{"$sample", bson.D{{"size", 1}}}}}
		cursor, _ := s.col.Aggregate(context.Background(), pipeline)

		var result Message
		for cursor.Next(context.Background()) {
			cursorErr := cursor.Decode(&result)
			if cursorErr != nil {
				log.Fatal("Error in random() cursor")
			}
		}

		if checkForVideo(result) {
			randomMessage(s)
		}

		result.Photos = handleMediaPath(result.Photos)
		jsonResult, _ := json.Marshal(result)

		w.Write(jsonResult)
	})
}

func capitalizeName(name string) string {
	return strings.Title(name)
}

func checkForVideo(obj Message) bool {

	fmt.Println("in check for video")
	if obj.Photos == nil {
		return false
	}

	path := obj.Photos[0].Uri
	ext := ".mp4"
	return strings.Contains(path, ext)
}

func handleMediaPath(origPhotos []Photo) []Photo {

	if origPhotos == nil {
		return origPhotos
	}

	if origPhotos[0].Uri == "" {
		return origPhotos
	}

	path := &origPhotos[0].Uri
	videos := "/videos/"
	photos := "/photos/"
	gifs := "/gifs/"

	if strings.Contains(*path, videos) {
		*path = stripVideoPath(*path)
	}

	if strings.Contains(*path, photos) {
		*path = stripPhotoPath(*path)
	}

	if strings.Contains(*path, gifs) {
		*path = stripGifPath(*path)
	}

	return origPhotos
}

func stripVideoPath(path string) string {
	videoIndex := strings.Index(path, "/videos/")
	return path[videoIndex:]
}

func stripPhotoPath(path string) string {
	splitString := strings.SplitAfter(path, "/photos/")
	return splitString[len(splitString)-1]
}

func stripGifPath(path string) string {
	gifIndex := strings.Index(path, "/gifs/")
	return path[gifIndex:]
}

func allMessagesBySender(s *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()["participant"]
		sender := capitalizeName(q[0])

		pipeline := []bson.D{
			bson.D{{"$match", bson.D{{"sender", sender}}}},
			bson.D{{"$sort", bson.D{{"timestamp", -1}}}}}
		cursor, _ := s.col.Aggregate(context.Background(), pipeline)

		var allMessages Messages

		for cursor.Next(context.Background()) {
			var result Message
			cursorErr := cursor.Decode(&result)
			if cursorErr != nil {
				log.Fatal("Error in random() cursor")
			}

			allMessages.Messages = append(allMessages.Messages, result)
		}

		jAllMessages, _ := json.Marshal(allMessages)
		w.Write(jAllMessages)
	})
}

func randomMessageBySender(s *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()["participant"]
		sender := capitalizeName(q[0])

		pipeline := []bson.D{
			bson.D{{"$match", bson.D{{"sender", sender}}}},
			bson.D{{"$sample", bson.D{{"size", 1}}}}}
		cursor, _ := s.col.Aggregate(context.Background(), pipeline)

		var allMessages Messages

		for cursor.Next(context.Background()) {
			var result Message
			cursorErr := cursor.Decode(&result)
			if cursorErr != nil {
				log.Fatal("Error in random() cursor")
			}

			allMessages.Messages = append(allMessages.Messages, result)
		}

		jAllMessages, _ := json.Marshal(allMessages)
		w.Write(jAllMessages)
	})
}

func getPort() string {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Google services wants 8080 or will decide for us.
		log.Printf("Defaulting to port %s", port)
	}

	return port
}

func main() {

	mongoURI := "mongodb://localhost:27017"
	//mongoURI := "mongodb+srv://kak:ricosuave@kak-6wzzo.gcp.mongodb.net/test?retryWrites=true&w=majority"
	//client, cancel := mongolib.ConnectToMongoDB(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))

	if err != nil {
		fmt.Println("Error connecting to mongo DB: ", err)
	}

	defer cancel()
	//collection := mongolib.GetMongoCollection(client, "nebrasketball", "messages")
	collection := client.Database("nebrasketball").Collection("messages")
	server := &Server{db: client, col: collection}

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.Handle("/random", randomMessage(server))
	http.Handle("/randsender", randomMessageBySender(server))
	http.Handle("/getallfromsender", allMessagesBySender(server))

	port := getPort()
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
