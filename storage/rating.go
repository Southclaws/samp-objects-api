package storage

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

// AddRating adds a rating to an object from a user
func (db *Database) AddRating(userID types.UserID, objectID types.ObjectID, value float64) (err error) {
	err = db.ratings.Insert(types.Rating{
		UserID:   userID,
		ObjectID: objectID,
		Value:    value,
		Date:     time.Now(),
	})
	if err != nil {
		return errors.Wrap(err, "failed to insert new rating")
	}
	fmt.Println(userID, objectID, value)
	err = db.objects.Update(
		bson.M{"id": objectID},
		bson.M{"$inc": bson.M{
			"ratecount": 1,
			"ratetotal": value,
		}})
	return errors.Wrap(err, "failed to increment rate count and rate total")
}
