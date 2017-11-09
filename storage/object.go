package storage

import (
	"errors"

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
	exists, err := db.ObjectExists(object)
	if err != nil {
		return
	}
	if exists {
		err = ErrObjectNameAlreadyExists
		return
	}

	if err = object.Validate(); err != nil {
		return
	}

	err = db.objects.Insert(object)

	return
}

// UpdateObject updates a object's account information
func (db Database) UpdateObject(object types.Object) (err error) {
	if err = object.Validate(); err != nil {
		return
	}

	err = db.objects.Update(bson.M{"id": object.ID}, object)

	return
}

// DeleteObject deletes a object's account
func (db Database) DeleteObject(object types.Object) (err error) {
	if err = object.Validate(); err != nil {
		return
	}

	err = db.objects.Remove(bson.M{"id": object.ID})

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

// GetObjectByName returns a types.Object by their name
func (db Database) GetObjectByName(name types.ObjectName) (object types.Object, err error) {
	err = db.objects.Find(bson.M{"name": name}).One(&object)
	return
}

// ObjectExists checks if a object exists by their unique ID
func (db Database) ObjectExists(object types.Object) (exists bool, err error) {
	count, err := db.users.Find([]bson.M{
		bson.M{"id": object.Owner},
		bson.M{"objects": []string{string(object.Name)}},
	}).Count()
	if err != nil {
		return
	}
	return count != 0, err
}
