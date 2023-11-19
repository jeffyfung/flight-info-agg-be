package model

import "time"

type Post struct {
	Title     string    `bson:"title" json:"title"`
	Summary   string    `bson:"summary" json:"summary"`
	Tags      []string  `bson:"tags" json:"tags"`
	URL       string    `bson:"url" json:"url"`
	PubDate   time.Time `bson:"pub_date" json:"pub_date"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
