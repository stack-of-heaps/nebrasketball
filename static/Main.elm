module NebrasketballMoms exposing (main)

import Http exposing (..)
import Html exposing (..)
import Html.Events exposing (onClick)
import Html.Attributes exposing (src)
import Json.Decode as Decode exposing (Decoder, decodeString, field, int, list, map3, string)
import Browser

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

type MessageType = TextMessage | PhotoMessage | ShareMessage | Unknown | Mixed

type alias MessengerMessage =
  {
    sender : String,
    content : String,
    timestamp: Int,
    photo : Maybe Photo,
    share : Maybe Share,
    reactions : Maybe (List Reaction),
    messageType: MessageType
  }

type alias Model = 
    {
        randomMessage : MessengerMessage,
        errorMessage : Maybe String
    }

view : Model -> Html Msg
view model = div [] 
    [
        h1 [] [ text "Random Message" ],
        viewMessageOrError model,
        button [ onClick SendHttpRequest ] [text "Random Message"]
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

messageIsLink : MessengerMessage -> Bool
messageIsLink message =
    if String.contains message.content "http" then True else False

renderMessageLink : MessengerMessage -> Html Msg
renderMessageLink message =
    div [] 
    [ 
        h4 [] [ text (message.sender ++ " sent a link") ],
        a [src message.content] [ text message.content ]
    ]

messageIsShare : MessengerMessage -> Bool
messageIsShare message = 
    case message.share of
       Nothing -> False
       Just link -> True

renderMessageShare : MessengerMessage -> Html Msg
renderMessageShare message =
    case message.share of
       Nothing -> Html.text ""
       Just share ->
            div [] 
            [ 
                h4 [] [ text (share.link ++ " sent a link") ],
                a [src message.content] [ text message.content ]
            ]

messageHasPhoto : MessengerMessage -> Bool
messageHasPhoto message = 
    case message.photo of
       Nothing -> False
       Just something -> if something.uri /= "" then True else False

renderMessagePhoto : MessengerMessage -> Html Msg
renderMessagePhoto message =
        case message.photo of
            Nothing -> Html.text ""
            Just aPhoto -> 
                let prependURL = "https://storage.cloud.google.com/directed-curve-263321.appspot.com/" in
                    if aPhoto.uri /= "" then (img [ src ( prependURL ++ aPhoto.uri ) ] [] ) else Html.text "Error -- Expected a photo URI here!"

messageIsOnlyText : MessengerMessage -> Bool
messageIsOnlyText message =
    if not (messageHasPhoto message) && not (messageIsShare message) then True else False

renderMessageText : MessengerMessage -> Html Msg
renderMessageText message = 
    div []
    [
        h4 [] [text (message.sender ++ ": " ++ message.content) ]
    ]

assignMessageType : MessengerMessage -> MessengerMessage
assignMessageType originalMessage =
    if messageIsShare originalMessage then { originalMessage | messageType = ShareMessage }
    else if messageHasPhoto originalMessage then { originalMessage | messageType = PhotoMessage }
    else if messageIsOnlyText originalMessage then { originalMessage | messageType = TextMessage }
    else { originalMessage | messageType = Mixed }

viewRandomMessage : MessengerMessage -> Html Msg
viewRandomMessage randomMessage =
    case randomMessage.messageType of 
        PhotoMessage -> renderMessagePhoto randomMessage
        ShareMessage -> renderMessageShare randomMessage
        TextMessage -> renderMessageText randomMessage
        _ ->
            div []
            [
                h2 [] [text "UnknownType"],
                h4 [] [text (randomMessage.sender ++ " : " ++ randomMessage.content) ]
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
    field "Timestamp" int

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

messageTypeDecoder : Decoder MessageType
messageTypeDecoder =
    Decode.succeed Unknown

messageDecoder : Decoder MessengerMessage
messageDecoder = 
  Decode.map7 MessengerMessage
    senderDecoder
    contentDecoder
    timestampDecoder
    photoDecoder
    shareDecoder
    reactionsDecoder
    messageTypeDecoder

update : Msg -> Model -> (Model, Cmd Msg)
update msg model = 
    case msg of
        SendHttpRequest -> (model, getRandomMessage)
        
        DataReceived (Ok messengerMessage) -> ( 
            let 
                messageWithType = assignMessageType messengerMessage 
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

initRandomMessage : MessengerMessage
initRandomMessage = 
    {
        sender = "",
        content = "",
        timestamp = 0,
        photo = Nothing,
        share = Nothing,
        reactions = Nothing,
        messageType = Unknown
    }

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