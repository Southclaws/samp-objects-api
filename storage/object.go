package storage

import (
	"errors"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

var (
	// ErrObjectNameAlreadyExists indicates that a object attempted to upload an object with a name
	// that they had already used
	ErrObjectNameAlreadyExists = errors.New("object already exists")
)

// CreateObject creates a new object in the database
func (db Database) CreateObject(object types.Object) (err error) {
	if err = object.Validate(); err != nil {
		return
	}

	err = db.objects.Insert(object)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE_OBJECT_NAME") {
			return ErrObjectNameAlreadyExists
		}
	}

	return
}

// UpdateObject updates a object's information
func (db Database) UpdateObject(object types.Object) (err error) {
	if err = object.Validate(); err != nil {
		return
	}

	err = db.objects.Update(bson.M{"id": object.ID}, object)
	return
}

// DeleteObject deletes a object
func (db Database) DeleteObject(object types.Object) (err error) {
	if err = object.Validate(); err != nil {
		return
	}

	err = db.objects.Remove(bson.M{"id": object.ID})
	if err != nil {
		return
	}

	return
}

// GetObject returns a types.Object by their unique ID
func (db Database) GetObject(id types.ObjectID) (object types.Object, err error) {
	err = db.objects.Find(bson.M{"id": id}).One(&object)
	if err != nil {
		return
	}

	return
}

// ObjectExists checks if a object exists by their unique ID
func (db Database) ObjectExists(objectID types.ObjectID) (exists bool, err error) {
	count, err := db.objects.Find(bson.M{"id": objectID}).Count()
	exists = count > 0
	return
}
