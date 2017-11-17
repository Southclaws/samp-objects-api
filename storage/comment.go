package storage

import (
	"time"

	"gopkg.in/mgo.v2/bson"
	"github.com/pkg/errors"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

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

	err = db.ratings.Insert(types.Comment{
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
