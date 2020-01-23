module NebrasketballMoms exposing (main)

import Http exposing (..)
import Html exposing (..)
import Html.Events exposing (onClick)
import Html.Attributes exposing (src, class, href)
import Json.Decode as Decode exposing (Decoder, decodeString, field, int, list, map3, string)
import Browser
import Css exposing (..)

randomUrl : String
randomUrl =
    "/random"

type alias Photo = { uri : String }

type alias Share = { link : String }

type alias Reaction =
 {
    reaction : String,
    actor : String
 }

type ProcessedMessage 
    = TextMessage { sender : String, content : String, timestamp: Int }
    | PhotoMessage { sender : String, content : String, timestamp: Int, photo: Photo }
    | ShareMessage { sender : String, content : String, timestamp: Int, share: Share }
    | EmptyMessage { sender: String, content: String, timestamp: Int }

type alias MessengerMessage =
  {
    sender : String,
    content : String,
    timestamp: Int,
    photo : Maybe Photo,
    share : Maybe Share,
    reactions : Maybe (List Reaction)
  }

type alias Model = 
    {
        randomMessage : ProcessedMessage,
        errorMessage : Maybe String
    }

view : Model -> Html Msg
view model = div [class "mainFlexDiv"] 
    [
        button [ onClick SendHttpRequest ] [text "Random Message"],
        div [class "subFlexDiv"]
        [
            viewMessageOrError model
        ]
    ]

viewError : String -> Html Msg
viewError error = 
    div []
    [
        h2 [] [text ("Error: " ++ error) ]
    ]

viewMessageOrError : Model -> Html Msg
viewMessageOrError model =
    case model.errorMessage of
        Just error -> 
            viewError error

        Nothing ->
            viewRandomMessage model.randomMessage

renderMessageLink : ProcessedMessage -> Html Msg
renderMessageLink message =
    case message of 
        TextMessage textMessage ->
            div [] 
            [ 
                h4 [] [ text (textMessage.sender ++ " sent a link") ],
                a [src textMessage.content] [ text textMessage.content ]
            ]
        _ -> text "" 

messageIsShare : MessengerMessage -> Bool
messageIsShare message = 
    case message.share of
       Nothing -> False
       Just link -> True

renderMessageShare : ProcessedMessage -> Html Msg
renderMessageShare message =
    case message of
        ShareMessage shareMessage ->
            div [] 
            [ 
                h4 [] [ text (shareMessage.share.link ++ " sent a link") ],
                a [src shareMessage.content] [ text shareMessage.content ]
            ]
        _ -> text "" 

messageHasPhoto : MessengerMessage -> Bool
messageHasPhoto message = 
    case message.photo of
       Nothing -> False
       Just something -> if something.uri /= "" then True else False

renderMessagePhoto : ProcessedMessage -> Html Msg
renderMessagePhoto message =
        case message of
            PhotoMessage photoMessage ->
                let 
                    prependURL = "https://storage.cloud.google.com/directed-curve-263321.appspot.com/"
                    imgURL = prependURL ++ photoMessage.photo.uri
                in
                    if 
                        photoMessage.photo.uri /= "" 
                    then 
                    div [] 
                        [
                            a [ href imgURL ] [img [src imgURL] []] 
                        ]
                    else 
                        Html.text "Error -- Expected a photo URI here!"
            _ -> text "" 

messageIsOnlyText : MessengerMessage -> Bool
messageIsOnlyText message =
    if not (messageHasPhoto message) && not (messageIsShare message) then True else False 

messageContainsLink : String -> Bool
messageContainsLink message =
    String.contains "http" message

renderMessageText : ProcessedMessage -> Html Msg
renderMessageText message = 
    case message of
        TextMessage textMessage ->
            if
                messageContainsLink textMessage.content
            then
                let 
                    messageWords = String.words textMessage.content
                    theLinkAsList = List.head (List.filter (\word -> String.contains "http" word) messageWords)
                    theLink = case theLinkAsList of
                        Nothing -> ""
                        Just aLink -> aLink
                    linkIndex = Maybe.withDefault 0 (List.head (String.indexes "http" textMessage.content))
                    messageLeftOfLink = String.slice 0 linkIndex textMessage.content
                    messageRightOfLink = String.slice (linkIndex + String.length theLink) (String.length textMessage.content) textMessage.content
                in
                    div [class "flexChildContainer"]
                    [
                        div[class "flexChild"] 
                        [
                            p [class "messageSender"] [text textMessage.sender],
                            p [class "messageText"] [text messageLeftOfLink],
                            a [href theLink] [text theLink],
                            p [class "messageText"] [text messageRightOfLink]
                        ],
                        div [class "flexChild"] 
                        [
                            iframe [class "customIframe", src theLink] []
                        ]
                    ]
            else
                div []
                [
                    p [class "messageSender"] [text textMessage.sender],
                    p [class "messageText"] [text textMessage.content]
                ]
        _ -> text "" 

processMessageType : MessengerMessage -> ProcessedMessage
processMessageType originalMessage =
    if messageIsShare originalMessage then makeShareMessage originalMessage
    else if messageHasPhoto originalMessage then makePhotoMessage originalMessage
    else if messageIsOnlyText originalMessage then makeTextMessage originalMessage
    else TextMessage { sender = "", content = "I've encountered something in the ProcessMessageType function which has fallen through into the base 'else' statement", timestamp = 0 }

makeShareMessage : MessengerMessage -> ProcessedMessage
makeShareMessage originalMessage =
    case originalMessage.share of
        Just aShare -> 
            ShareMessage { sender = originalMessage.sender, content = originalMessage.content, timestamp = originalMessage.timestamp, share = aShare }
        Nothing -> 
            ShareMessage { sender = originalMessage.sender, content = originalMessage.content, timestamp = originalMessage.timestamp, share = Share "Whoops! There was an error processing this Share Message." }

makePhotoMessage : MessengerMessage -> ProcessedMessage
makePhotoMessage originalMessage =
    case originalMessage.photo of
        Just aPhoto -> 
            PhotoMessage { sender = originalMessage.sender, content = originalMessage.content, timestamp = originalMessage.timestamp, photo = aPhoto }
        Nothing ->
            PhotoMessage { sender = originalMessage.sender, content = originalMessage.content, timestamp = originalMessage.timestamp, photo = Photo "Whoops! There was an error processing this Photo Message." } 


makeTextMessage : MessengerMessage -> ProcessedMessage
makeTextMessage originalMessage =
    TextMessage { sender = originalMessage.sender, content = originalMessage.content, timestamp = originalMessage.timestamp }

viewRandomMessage : ProcessedMessage -> Html Msg
viewRandomMessage randomMessage =
    case randomMessage of 
        PhotoMessage photoMessage -> renderMessagePhoto randomMessage
        ShareMessage shareMessage -> renderMessageShare randomMessage
        TextMessage textMessage -> renderMessageText randomMessage
        _ ->
            div []
            [
            ]

type Msg
    = SendHttpRequest
    | DataReceived (Result Http.Error MessengerMessage)

getRandomMessage : Cmd Msg
getRandomMessage =
    Http.get 
    { url = randomUrl,
    expect = Http.expectJson DataReceived messageDecoder
    }

senderDecoder : Decoder String
senderDecoder =
    field "Sender" string

timestampDecoder : Decoder Int
timestampDecoder =
    field "Timestamp" Decode.int

contentDecoder : Decoder String
contentDecoder =
    field "Content" string

photoDecoder : Decoder ( Maybe Photo )
photoDecoder =
        Decode.maybe ( 
            Decode.map Photo ( 
                field "Photo" ( field "Uri" string ) ) )

shareDecoder : Decoder ( Maybe Share )
shareDecoder =
    Decode.maybe ( 
        Decode.map Share
            ( field "Link" string ) )

reactionDecoder : Decoder Reaction
reactionDecoder =
    Decode.map2 Reaction
    (field "Reaction" string)
    (field "Actor" string)

reactionsDecoder : Decoder ( Maybe ( List Reaction ))
reactionsDecoder = 
    Decode.maybe (Decode.list reactionDecoder)

messageDecoder : Decoder MessengerMessage
messageDecoder = 
  Decode.map6 MessengerMessage
    senderDecoder
    contentDecoder
    timestampDecoder
    photoDecoder
    shareDecoder
    reactionsDecoder

update : Msg -> Model -> (Model, Cmd Msg)
update msg model = 
    case msg of
        SendHttpRequest -> (model, getRandomMessage)
        
        DataReceived (Ok messengerMessage) -> ( 
            let 
                messageWithType = processMessageType messengerMessage 
            in
                {model | randomMessage = messageWithType, errorMessage = Nothing }, Cmd.none )
        
        DataReceived (Err httpError) -> ( { model | errorMessage = Just (buildErrorMessage httpError )}, Cmd.none )

buildErrorMessage : Http.Error -> String
buildErrorMessage httpError = 
    case httpError of
        Http.BadUrl message ->
            message

        Http.Timeout -> 
            "Connection to server or database timed out."

        Http.NetworkError ->
            "Network error."

        Http.BadStatus statusCode ->
            "Server responded with bad status: " ++ String.fromInt statusCode

        Http.BadBody message ->
            message

initRandomMessage : ProcessedMessage
initRandomMessage = EmptyMessage {sender = "", content = " ", timestamp = 0 }

init : () -> ( Model, Cmd Msg )
init _ =
    ( {errorMessage = Nothing,
    randomMessage = initRandomMessage
      }
    , Cmd.none
    )

main : Program () Model Msg
main =
    Browser.element
        { init = init
        , view = view
        , update = update
        , subscriptions = \_ -> Sub.none
        }