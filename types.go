package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Participants struct {
	Name string
}

type Reaction struct {
	Reaction string `json:"reaction"`
	Actor    string `json:"actor"`
}

type Photo struct {
	Uri string `json:"uri"`
}

type Video struct {
	Uri string `json:"uri"`
}

type Gif struct {
	Uri string `json:"uri"`
}

type Share struct {
	Link      string `json:"link"`
	ShareText string `bson:"share_text" json:"shareText"`
}

type Messages struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Id        primitive.ObjectID `bson:"_id" json:"id"`
	Sender    string             `bson:"sender_name" json:"sender"`
	Timestamp int64              `bson:"timestamp_ms" json:"timestamp"`
	Content   string             `json:"content"`
	Photos    []Photo            `json:"photos"`
	Reactions []Reaction         `json:"reactions"`
	Gifs      []Gif              `json:"gifs"`
	Videos    []Video            `json:"videos"`
	Share     Share              `json:"share"`
	Type      string             `json:type"`
}
