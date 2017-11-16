package types

import (
	"time"
)

// Rating represents a user's 1-5 star rating on an object
type Rating struct {
	UserID   UserID    `json:"user"`
	ObjectID ObjectID  `json:"object"`
	Value    float64   `json:"value"`
	Date     time.Time `json:"date"`
}
