package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"go.mongodb.org/mongo-driver/bson"
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

type ReturnMessage struct {
	Sender    string
	Timestamp int
	Content   string
	Photo     Photo
	Reactions []Reaction
	Share     Share
	Type      string
}

type ServerResponse struct {
	MessageResults Messages
	Error          ErrorCode
	LastID         string
}

type ErrorCode string

const (
	KeyPassPhrase             string    = "fjklj4kj12414980a9fasdvklavn!@$1"
	MalformedPagedBySenderURL ErrorCode = "URL should look like '...?sender=example%20name&startAt=a8890ef6b...'"
	SenderEmpty               ErrorCode = "URL 'sender' parameter is empty"
)

// ObjectIdRegEx Only grabs alphanumeric ID and quotes between ObjectID()
// TODO: BETTER WAY TO DO THIS?
var ObjectIdRegEx = regexp.MustCompile(`"(.*?)"`)

func randomMessage(s *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Begin randomMessage")

		// Construct aggregation "pipeline" to return 1 random document from entire collection
		pipeline := []bson.D{bson.D{{"$sample", bson.D{{"size", 1}}}}}
		cursor, _ := s.col.Aggregate(context.Background(), pipeline)

		var result Message
		for cursor.Next(context.Background()) {
			cursorErr := cursor.Decode(&result)
			if cursorErr != nil {
				log.Fatal("Error in random() cursor")
			}

			fmt.Println("Result: ", result)
		}

		if checkForVideo(result) {
			randomMessage(s)
		}

		retMessage := craftReturnMessage(result)

		jsonResult, _ := json.Marshal(retMessage)

		w.Write(jsonResult)

		fmt.Println("End randomMessage")
	})
}

func createEmptyServerResponseWithError(err ErrorCode) ServerResponse {

	return ServerResponse{
		Error:          err,
		MessageResults: Messages{},
		LastID:         ""}
}

// First string 	= sender
// Second string 	= startingId (if any)
// If ServerResponse != nil -> Return it, because we have an error

func getPagedQueryTerms(r *http.Request) (sender string, startingId string, response ServerResponse) {

	query := r.URL.Query()

	if len(query) == 0 {
		responseObject := createEmptyServerResponseWithError(MalformedPagedBySenderURL)
		return "", "", responseObject
	}

	senderQueryParam := query["sender"]
	if len(senderQueryParam) == 0 {

		responseObject := createEmptyServerResponseWithError(SenderEmpty)
		return "", "", responseObject
	}

	sender = senderQueryParam[0]

	if sender == "" {
		responseObject := createEmptyServerResponseWithError(SenderEmpty)
		return "", "", responseObject
	}

	startingIdQ := query["startAt"]
	if len(startingIdQ) == 0 {
		startingId = ""
	} else {
		startingId = startingIdQ[0]
	}

	return sender, startingId, ServerResponse{}
}

func encryptLastId(lastId string) string {

	fmt.Println("Beginning encryptLastId()")

	// Generate AES cipher with 32 byte passphrase
	aesCipher, err := aes.NewCipher([]byte(KeyPassPhrase))

	if err != nil {
		fmt.Println("Error in encryptLastId(): ", err)
	}

	// GCM "Galois/Counter Mode": Symmetric Keyy cryptographic block cipher
	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		fmt.Println("Error in encryptLastId(): ", err)
	}

	// Nonce is literally a "one off" byte array which will be populated by a random sequence below.
	// The nonce is prepended/appended to the cipher (?) and is used in deciphering
	nonce := make([]byte, gcm.NonceSize())

	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println("Error in io.ReadFull: ", err)
	}

	encryptedByteArray := gcm.Seal(nonce, nonce, []byte(lastId), nil)

	// Convert to Base64 to ensure we can transmit via HTTP without error or corruption
	encryptedString := base64.StdEncoding.EncodeToString(encryptedByteArray)

	fmt.Println("Ending encryptLastId()")

	return encryptedString
}

func decryptLastId(encLastId string) string {

	fmt.Println("Beginning decryptLastId()")

	encLastIdByteArray, err := base64.StdEncoding.DecodeString(encLastId)

	if err != nil {
		fmt.Println("Error in StdEncoding.DecodeString: ", err)
	}

	aesCipher, err := aes.NewCipher([]byte(KeyPassPhrase))

	if err != nil {
		fmt.Println("Error in decryptLastId(): ", err)
	}

	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		fmt.Println("Error in encryptLastId(): ", err)
	}

	nonceSize := gcm.NonceSize()

	nonce, cipherText := encLastIdByteArray[:nonceSize], encLastIdByteArray[nonceSize:]

	decryptedLastId, err := gcm.Open(nil, []byte(nonce), []byte(cipherText), nil)

	if err != nil {
		fmt.Println("Error in gcm.Open: ", err)
	}

	fmt.Println("Ending decryptLastId()")

	return string(decryptedLastId)

}

func pagedMessagesLogic(s *Server, r *http.Request) ServerResponse {

	fmt.Println("Begin pagedMessagesBySender()")

	maxItems := 10

	sender, startingId, err := getPagedQueryTerms(r)

	if err.Error != "" {
		return err
	}

	fmt.Println("StartingID: ", startingId)
	fmt.Println("Sender: ", sender)

	// pipeline := []bson.D{bson.D{{"$match", {bson.D{{"sender", sender}}}}}, bson.D{{"$limit", maxItems}}}
	// pipeline := []bson.M{bson.M{"$match": bson.M{"sender": sender, "_id": bson.M{"$gt": startingId}}}, bson.M{"$limit": maxItems}}

	pipeline := pagedPipelineBuilder(sender, startingId, maxItems)
	cursor, _ := s.col.Aggregate(context.Background(), pipeline)

	var messageBatch Messages
	var result Message
	var rawId bson.RawValue

	for cursor.Next(context.Background()) {
		cursorErr := cursor.Decode(&result)
		if cursorErr != nil {
			log.Fatal("Error in pagedMessagesBySender() cursor: ", cursorErr)
		}

		messageBatch.Messages = append(messageBatch.Messages, result)
		rawId = cursor.Current.Lookup("_id")
	}

	lastId := stringFromRawValue(rawId)
	encryptedLastId := encryptLastId(lastId)

	serverResponse := ServerResponse{
		MessageResults: messageBatch,
		Error:          "",
		LastID:         encryptedLastId}

	return serverResponse
}

func stringFromRawValue(rawId bson.RawValue) string {
	objectID := rawId.ObjectID().String()
	lastId := strings.Split(objectID, "\"")

	return lastId[1]
}

func pagedPipelineBuilder(sender string, startingId string, limit int) []bson.M {

	//pipeline := []bson.M{bson.M{"$match": bson.M{"sender": sender, "_id": bson.M{"$gt": startingId}}}, bson.M{"$limit": maxItems}}
	matchElement := matchPipelineBuilder(sender, startingId)
	limitElement := bson.M{"$limit": limit}
	pipeline := []bson.M{matchElement, limitElement}

	return pipeline
}

func matchPipelineBuilder(sender string, startingId string) bson.M {

	matchRoot := bson.M{"$match": ""}
	senderElement := bson.M{"sender": sender}
	idElement := bson.M{"_id": ""}
	gtElement := bson.M{"$gt": ""}

	if startingId == "" {
		matchRoot["$match"] = senderElement
	} else {
		gtElement["$gt"] = startingId
		idElement["_id"] = gtElement
		tempArray := []bson.M{senderElement, idElement}
		matchRoot["$match"] = tempArray
	}

	fmt.Println("REturning matchroot: ", matchRoot)
	return matchRoot
}

func pagedMessagesBySender(s *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		returnObject := pagedMessagesLogic(s, r)

		returnJson, err := json.Marshal(returnObject)

		if err != nil {
			fmt.Println("Error converted pagedMessagesLogic() response to JSON: ", err)
		}

		w.Write(returnJson)
	})
}

func craftReturnMessage(objIn Message) ReturnMessage {

	objIn.Photos = handleMediaPath(objIn.Photos)

	newMessage := ReturnMessage{
		Sender:    objIn.Sender,
		Content:   objIn.Content,
		Timestamp: objIn.Timestamp,
		Share:     objIn.Share,
		Reactions: objIn.Reactions,
	}

	if len(objIn.Photos) > 0 {
		newMessage.Photo = objIn.Photos[0]
	}

	return newMessage
}

func capitalizeName(name string) string {
	return cases.Title(language.English).String(name)
}

func checkForVideo(obj Message) bool {

	if obj.Photos == nil {
		return false
	}

	path := obj.Photos[0].Uri
	ext := ".mp4"
	fmt.Println("Path: ", path)
	return strings.Contains(path, ext)
}

func handleMediaPath(origPhotos []Photo) []Photo {

	if origPhotos == nil || origPhotos[0].Uri == "" {
		return origPhotos
	}

	path := &origPhotos[0].Uri
	videos := "/videos/"
	photos := "/photos/"
	gifs := "/gifs/"

	if strings.Contains(*path, videos) {
		*path = stripPath(*path, videos)
	}

	if strings.Contains(*path, photos) {
		*path = stripPath(*path, photos)
	}

	if strings.Contains(*path, gifs) {
		*path = stripPath(*path, gifs)
	}

	return origPhotos
}

func stripPath(path string, substringToStrip string) string {
	pathIndex := strings.Index(path, substringToStrip)
	return path[pathIndex:]
}

func allMessagesBySender(s *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()["participant"]
		sender := capitalizeName(q[0])

		pipeline := []bson.D{
			{{"$match", bson.D{{"sender", sender}}}},
			{{"$sort", bson.D{{"timestamp", -1}}}}}
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

	// TODO: SET VIA CONFIGURATION
	//mongoURI := "mongodb://localhost:27017"
	mongoURI := "mongodb+srv://kak:ricosuave@kak-6wzzo.gcp.mongodb.net/test?retryWrites=true&w=majority"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))

	if err != nil {
		fmt.Println("Error connecting to mongo DB: ", err)
	}

	defer cancel()
	collection := client.Database("nebrasketball").Collection("messages")
	server := &Server{db: client, col: collection}

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.Handle("/random", randomMessage(server))
	http.Handle("/randsender", randomMessageBySender(server))
	http.Handle("/getallfromsender", allMessagesBySender(server))
	http.Handle("/getpagedfromsender", pagedMessagesBySender(server))

	port := getPort()
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
