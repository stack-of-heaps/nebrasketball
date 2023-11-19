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
	"nebrasketball/dbAccessor"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ReturnMessage struct {
	Sender    string
	Timestamp int
	Content   string
	Photo     dbAccessor.Photo
	Reactions []dbAccessor.Reaction
	Share     dbAccessor.Share
	Type      string
}

type ServerResponse struct {
	MessageResults dbAccessor.Messages
	Error          string
	LastID         string
}

const (
	KeyPassPhrase             string = "fjklj4kj12414980a9fasdvklavn!@$1"
	MalformedPagedBySenderURL string = "URL should look like '...?sender=example%20name&startAt=a8890ef6b...'"
	SenderEmpty               string = "URL 'sender' parameter is empty"
)

func createEmptyServerResponseWithError(err string) ServerResponse {

	return ServerResponse{
		Error:          err,
		MessageResults: dbAccessor.Messages{},
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

func stringFromRawValue(rawId bson.RawValue) string {
	objectID := rawId.ObjectID().String()
	lastId := strings.Split(objectID, "\"")

	return lastId[1]
}

func pagedMessagesBySender(s *MongoClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		returnObject := pagedMessagesLogic(s, r)

		returnJson, err := json.Marshal(returnObject)

		if err != nil {
			fmt.Println("Error converted pagedMessagesLogic() response to JSON: ", err)
		}

		w.Write(returnJson)
	})
}

func craftReturnMessage(objIn dbAccessor.Message) ReturnMessage {

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

		var message = dbAccessor.getMessages(s, "participant", q, true)

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
