package storage

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go"
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
func (db Database) PutObjectFile(objectID types.ObjectID, filename string, filesize int64, reader io.Reader) (err error) {
	if err = objectID.Validate(); err != nil {
		return
	}

	_, err = db.store.PutObject(
		db.StoreBucket,
		filepath.Join(string(objectID), filename),
		reader,
		filesize,
		minio.PutObjectOptions{})

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

// GetObjectFiles returns an object's associated files in memory
func (db Database) GetObjectFiles(objectID types.ObjectID) (objectFiles types.ObjectFiles, err error) {
	if err = objectID.Validate(); err != nil {
		return
	}

	doneCh := make(chan struct{})
	infoCh := db.store.ListObjects(db.StoreBucket, string(objectID), true, doneCh)
	var (
		file *minio.Object
		stat minio.ObjectInfo
		n    int
	)
	for info := range infoCh {
		ext := filepath.Ext(info.Key)
		if ext == ".dff" {
			objectDFF := types.ObjectDFF{}
			file, err = db.store.GetObject(db.StoreBucket, info.Key, minio.GetObjectOptions{})
			if err != nil {
				err = errors.Wrap(err, "failed to get object info")
				return
			}
			defer file.Close() // nolint

			stat, err = file.Stat()
			if err != nil {
				err = errors.Wrap(err, "failed to stat file")
				return
			}

			objectDFF.Data = make([]byte, stat.Size)
			n, err = file.Read(objectDFF.Data)
			if err != nil && !(err == io.EOF && int64(n) == stat.Size) {
				err = errors.Wrap(err, "failed to read byte stream")
				return
			}
			err = nil

			objectDFF.Name = filepath.Base(info.Key)
			objectFiles.Models = append(objectFiles.Models, objectDFF)
		} else if ext == ".txd" {
			objectTXD := types.ObjectTXD{}
			file, err = db.store.GetObject(db.StoreBucket, info.Key, minio.GetObjectOptions{})
			if err != nil {
				err = errors.Wrap(err, "failed to get object info")
				return
			}
			defer file.Close() // nolint

			stat, err = file.Stat()
			if err != nil {
				err = errors.Wrap(err, "failed to stat file")
				return
			}

			objectTXD.Data = make([]byte, stat.Size)
			n, err = file.Read(objectTXD.Data)
			if err != nil && !(err == io.EOF && int64(n) == stat.Size) {
				err = errors.Wrap(err, "failed to read byte stream")
				return
			}
			err = nil

			objectTXD.Name = filepath.Base(info.Key)
			objectFiles.Textures = append(objectFiles.Textures, objectTXD)
		}
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
	count, err := db.objects.Find(bson.M{"owner": object.Owner, "name": object.Name}).Count()
	exists = count > 0
	return
}
