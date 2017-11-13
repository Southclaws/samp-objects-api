package storage

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
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

// PutObjectFile uploads a file to an object's folder in S3 from an io.Reader
func (db Database) PutObjectFile(objectID types.ObjectID, filename string, reader io.Reader) (err error) {
	if err = objectID.Validate(); err != nil {
		return
	}

	_, err = db.store.PutObject(
		db.StoreBucket,
		filepath.Join(string(objectID), filename),
		reader,
		"application/octet-stream")

	return
}

// GetObjectThumb writes the first image from an object to the given writer
func (db Database) GetObjectThumb(objectID types.ObjectID, writer io.Writer) (err error) {
	if err = objectID.Validate(); err != nil {
		err = errors.Wrap(err, "invalid object ID format")
		return
	}

	tmpObject := types.Object{}
	err = db.objects.Find(bson.M{"id": objectID}).One(&tmpObject)
	if err != nil {
		err = errors.Wrapf(err, "failed to lookup object %s", string(objectID))
		return
	}

	storeObject, err := db.store.GetObject(
		db.StoreBucket,
		filepath.Join(string(objectID), string(tmpObject.Images[0])))
	if err != nil {
		err = errors.Wrap(err, "failed to get file from object store")
		return
	}

	_, err = io.Copy(writer, storeObject)

	return
}

// GetObjectFile writes the specified image file from an object to the given writer
func (db Database) GetObjectFile(objectID types.ObjectID, fileName types.File, writer io.Writer) (err error) {
	if err = objectID.Validate(); err != nil {
		err = errors.Wrap(err, "invalid object ID format")
		return
	}

	storeObject, err := db.store.GetObject(
		db.StoreBucket,
		filepath.Join(string(objectID), string(fileName)))
	if err != nil {
		err = errors.Wrap(err, "failed to get file from object store")
		return
	}

	_, err = io.Copy(writer, storeObject)

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

// GetObjects returns a list of objects based on query parameters
func (db Database) GetObjects( /*todo: query params*/ ) (objects []types.Object, err error) {
	err = db.objects.Find(bson.M{}).All(&objects)
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

// GetUserObject returns a types.Object from a specific owner and an object name
func (db Database) GetUserObject(userName types.UserName, objectName types.ObjectName) (object types.Object, err error) {
	if err = userName.Validate(); err != nil {
		return
	}
	// if err = objectName.Validate(); err != nil {
	// 	return
	// }

	err = db.objects.Find(bson.M{"name": objectName, "ownername": userName}).One(&object)
	if err != nil {
		return
	}

	return
}

// ObjectExists checks if an object exists by their unique ID
func (db Database) ObjectExists(objectID types.ObjectID) (exists bool, err error) {
	count, err := db.objects.Find(bson.M{"id": objectID}).Count()
	exists = count > 0
	return
}

// UserObjectExists checks if an object exists by their name in a user's account
func (db Database) UserObjectExists(object types.Object) (exists bool, err error) {
	count, err := db.objects.Find(bson.M{"ownerid": object.OwnerID, "name": object.Name}).Count()
	exists = count > 0
	return
}
