module HttpExamples exposing (main)

import Http exposing (..)
import Html exposing (..)
import Html.Events exposing (onClick)
import Html.Attributes exposing (src)
import Json.Decode exposing (Decoder, decodeString, field, int, list, map3, string)
import Browser

type alias MessengerMessage = 
    {
        sender : String,
        timestamp : Int,
        content : String
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
        h4 [] [text (randomMessage.sender ++ ":" ++ randomMessage.content) ],
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
    expect = Http.expectJson DataReceived messengerMessageDecoder
    }

messengerMessageDecoder : Decoder MessengerMessage
messengerMessageDecoder =
    map3 MessengerMessage
        (field "Sender" string)
        (field "Timestamp" int)
        (field "Content" string)

update : Msg -> Model -> (Model, Cmd Msg)
update msg model = 
    case msg of
        SendHttpRequest -> (model, getRandomMessage)
        DataReceived (Ok messengerMessage) -> 
            ( {model | randomMessage = messengerMessage, errorMessage = Nothing }, Cmd.none )
        DataReceived (Err httpError) -> 
            (
                { model | errorMessage = Just (buildErrorMessage httpError )}, Cmd.none 
            )

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
        timestamp = 0,
        content = ""
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