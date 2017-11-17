package types

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Comment represents a user comment on an object
type Comment struct {
	ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	UserID   UserID        `json:"user"`
	ObjectID ObjectID      `json:"object"`
	Content  string        `json:"content"`
	Date     time.Time     `json:"date"`
}
