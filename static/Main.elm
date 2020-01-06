module HttpExamples exposing (main)

import Http exposing (..)
import Html exposing (..)
import Html.Events exposing (onClick)
import Html.Attributes exposing (src)
import Json.Decode as Decode exposing (Decoder, decodeString, field, int, list, map3, string)
import Browser

type alias Photo =
  { 
    uri : String
  }

type alias Reaction =
 {
    reaction : String,
    actor : String
 }

type alias Share =
  {
    link : String
  }

type alias MessengerMessage =
  {
    sender : String,
    content : String,
    timestamp: Int,
    photos : Maybe (List String),
    shares : Maybe (List String),
    reactions : Maybe (List Reaction)
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
        viewMessageOrError model
    ]

viewMessageOrError : Model -> Html Msg
viewMessageOrError model =
    case model.errorMessage of
        Just error -> 
            viewError error

        Nothing ->
            viewRandomMessage model.randomMessage

viewRandomMessage : MessengerMessage -> Html Msg
viewRandomMessage randomMessage =
    div []
    [
        h2 [] [text "Header"],
        h4 [] [text (randomMessage.sender ++ " : " ++ randomMessage.content) ],
        button [ onClick SendHttpRequest ]
            [text "Random Message"]
    ]

viewError : String -> Html Msg
viewError error = 
    div []
    [
        h2 [] [text ("Error: " ++ error) ]
    ]

type Msg
    = SendHttpRequest
    | DataReceived (Result Http.Error MessengerMessage)

randomUrl : String
randomUrl =
    "/random"

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

photoDecoder : Decoder String
photoDecoder =
    (field "Photos" (field "Uri" string))

photosDecoder : Decoder (Maybe (List String))
photosDecoder =
  Decode.maybe (list photoDecoder)

reactionDecoder : Decoder Reaction
reactionDecoder =
    Decode.map2 Reaction
    (field "Reaction" string)
    (field "Actor" string)

reactionsDecoder : Decoder (Maybe (List Reaction))
reactionsDecoder =
  Decode.maybe (Decode.list reactionDecoder)

shareDecoder : Decoder String 
shareDecoder =
    (field "link" string)

sharesDecoder : Decoder (Maybe (List String))
sharesDecoder =
  Decode.maybe (Decode.list shareDecoder)



messageDecoder : Decoder MessengerMessage
messageDecoder = 
  Decode.map6 MessengerMessage
    senderDecoder
    contentDecoder
    timestampDecoder
    photosDecoder
    sharesDecoder
    reactionsDecoder

update : Msg -> Model -> (Model, Cmd Msg)
update msg model = 
    case msg of
        SendHttpRequest -> (model, getRandomMessage)
        
        DataReceived (Ok messengerMessage) -> ( {model | randomMessage = messengerMessage, errorMessage = Nothing }, Cmd.none )
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


initReaction : Reaction
initReaction = 
    {
        reaction = "",
        actor = ""
    }

initReactions : (List Reaction)
initReactions = [initReaction]
initRandomMessage : MessengerMessage
initRandomMessage = 
    {
        sender = "",
        content = "",
        timestamp = 0,
        photos = Nothing,
        shares = Nothing,
        reactions = Nothing
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