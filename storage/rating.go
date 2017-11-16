package storage

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

// AddRating adds a rating to an object from a user
func (db *Database) AddRating(userID types.UserID, objectID types.ObjectID, value float64) (exists bool, err error) {
	if err = userID.Validate(); err != nil {
		return
	}
	if err = objectID.Validate(); err != nil {
		return
	}
	if value < 0.0 || value > 5.0 {
		return false, errors.New("invalid rating value")
	}

	err = db.ratings.Insert(types.Rating{
		UserID:   userID,
		ObjectID: objectID,
		Value:    value,
		Date:     time.Now(),
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE_USER_OBJECT_RATING") {
			return true, nil
		}
		return false, errors.Wrap(err, "failed to insert new rating")
	}

	err = db.objects.Update(
		bson.M{"id": objectID},
		bson.M{"$inc": bson.M{
			"ratecount": 1,
			"ratetotal": value,
		}})
	return
}

// RemoveRating removes a user's rating from an object
func (db *Database) RemoveRating(userID types.UserID, objectID types.ObjectID) (err error) {
	if err = userID.Validate(); err != nil {
		return
	}
	if err = objectID.Validate(); err != nil {
		return
	}

	rating := types.Rating{}
	err = db.ratings.Find(bson.M{"userid": userID, "objectid": objectID}).One(&rating)
	if err != nil {
		return errors.Wrap(err, "failed to find rating")
	}

	err = db.ratings.Remove(bson.M{"userid": userID, "objectid": objectID})
	if err != nil {
		return errors.Wrap(err, "failed to remove rating")
	}

	err = db.objects.Update(
		bson.M{"id": objectID},
		bson.M{"$inc": bson.M{
			"ratecount": -1,
			"ratetotal": -rating.Value,
		}})
	return
}
