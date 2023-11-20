package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

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
	Error          string
	LastID         string
}

const (
	MalformedPagedBySenderURL string = "URL should look like '...?sender=example%20name&startAt=a8890ef6b...'"
	SenderEmpty               string = "URL 'sender' parameter is empty"
)

func createEmptyServerResponseWithError(err string) ServerResponse {

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

func pagedMessagesBySender(s *MongoClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sender, startingId, errorResponse := getPagedQueryTerms(r)

		if errorResponse.Error != "" {
			w.Write([]byte(errorResponse.Error))
		}

		messages, encryptedLastId := pagedMessagesLogic(s, sender, startingId)

		response := ServerResponse{
			Error:          "",
			MessageResults: messages,
			LastID:         encryptedLastId}

		returnJson, err := json.Marshal(response)

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

func randomMessageBySender(s *MongoClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()["participant"]
		sender := capitalizeName(q[0])
		message := getMessages(s, sender, "participant", true)

		jAllMessages, _ := json.Marshal(message)
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
	server := &MongoClient{db: client, col: collection}

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
