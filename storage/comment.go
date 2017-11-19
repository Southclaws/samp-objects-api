package storage

import (
	"time"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"

	"github.com/Southclaws/samp-objects-api/types"
)

// GetComments returns a slice of types.Comment for a given object ID
func (db *Database) GetComments(objectID types.ObjectID) (comments []types.Comment, err error) {
	if err = objectID.Validate(); err != nil {
		return
	}

	err = db.comments.Find(bson.M{"objectid": objectID}).All(&comments)
	return
}

// AddComment creates a comment from a user on an object
func (db *Database) AddComment(userID types.UserID, objectID types.ObjectID, content string) (err error) {
	if err = userID.Validate(); err != nil {
		return
	}
	if err = objectID.Validate(); err != nil {
		return
	}
	if len(content) > 1024 {
		return errors.New("content too large")
	}

	err = db.comments.Insert(types.Comment{
		UserID:   userID,
		ObjectID: objectID,
		Content:  content,
		Date:     time.Now(),
	})
	if err != nil {
		return errors.Wrap(err, "failed to insert new comment")
	}

	return
}

// RemoveComment removes a comment by ID (MongoDB's automatically assigned ObjectID)
func (db *Database) RemoveComment(commentID bson.ObjectId) (err error) {
	err = db.comments.Remove(bson.M{"_id": commentID})
	return
}
