package storage

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go"

	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

var (
	// ErrObjectNameAlreadyExists indicates that a object attempted to upload an object with a name
	// that they had already used
	ErrObjectNameAlreadyExists = errors.New("object already exists")
)

// CreateObject creates a new object in the database
func (db Database) CreateObject(object types.Object, objectData types.ObjectFiles) (err error) {
	if err = object.Validate(); err != nil {
		return
	}

	for _, model := range objectData.Models {
		objectReader := bytes.NewReader(model.Data)
		_, err = db.store.PutObject(
			db.StoreBucket,
			filepath.Join(string(object.ID), model.Name),
			objectReader,
			int64(-1),
			minio.PutObjectOptions{})
		if err != nil {
			return
		}
	}

	for _, texture := range objectData.Textures {
		objectReader := bytes.NewReader(texture.Data)
		_, err = db.store.PutObject(
			db.StoreBucket,
			filepath.Join(string(object.ID), texture.Name),
			objectReader,
			int64(-1),
			minio.PutObjectOptions{})
		if err != nil {
			return
		}
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
func (db Database) DeleteObject(objectID types.ObjectID) (err error) {
	if err = objectID.Validate(); err != nil {
		return
	}

	doneCh := make(chan struct{})
	infoCh := db.store.ListObjects(db.StoreBucket, string(objectID), true, doneCh)
	for object := range infoCh {
		err = db.store.RemoveObject(db.StoreBucket, object.Key)
		if err != nil {
			return
		}
	}

	err = db.objects.Remove(bson.M{"id": objectID})
	if err != nil {
		return
	}

	return
}

// GetObject returns a types.Object by their unique ID
func (db Database) GetObject(objectID types.ObjectID) (object types.Object, err error) {
	err = db.objects.Find(bson.M{"id": objectID}).One(&object)
	if err != nil {
		return
	}

	return
}

// GetObjectFiles returns an object's associated files in memory
func (db Database) GetObjectFiles(objectID types.ObjectID) (objectFiles types.ObjectFiles) {
	return
}

// ObjectExists checks if a object exists by their unique ID
func (db Database) ObjectExists(objectID types.ObjectID) (exists bool, err error) {
	count, err := db.objects.Find(bson.M{"id": objectID}).Count()
	exists = count > 0
	return
}
