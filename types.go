package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Participants struct {
	Name string
}

type Reaction struct {
	Reaction string
	Actor    string
}

type Photo struct {
	Uri string
}

type Video struct {
	Uri string
}

type Gif struct {
	Uri string
}

type Share struct {
	Link      string
	ShareText string `bson:"share_text"`
}

type Messages struct {
	Messages []Message
}

type Message struct {
	Id        primitive.ObjectID `bson:"_id"`
	Sender    string             `bson:"sender_name"`
	Timestamp int64              `bson:"timestamp_ms"`
	Content   string
	Photos    []Photo
	Reactions []Reaction
	Gifs      []Gif
	Videos    []Video
	Share     Share
	Type      string
}
