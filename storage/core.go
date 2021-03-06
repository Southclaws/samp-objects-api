package storage

import (
	"fmt"

	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
)

// Database represents the storage backend state
type Database struct {
	session  *mgo.Session
	users    *mgo.Collection
	objects  *mgo.Collection
	ratings  *mgo.Collection
	comments *mgo.Collection
	store    *minio.Client

	StoreBucket   string
	StoreLocation string
}

// Config represents the configuration required to interact with the database
type Config struct {
	MongoHost           string
	MongoPort           string
	MongoUser           string
	MongoPass           string
	MongoName           string
	MongoCollectionInfo mgo.CollectionInfo
	StoreHost           string
	StorePort           string
	StoreAccess         string
	StoreSecret         string
	StoreSecure         bool
	StoreBucket         string
	StoreLocation       string
}

// New simply provides a function to set up a MongoDB connection and perform some checks
// against the selected database/collection to ensure it's ready for use.
func New(config Config) (*Database, error) {
	var (
		err      error
		database Database
	)

	database.session, err = mgo.Dial(fmt.Sprintf("%s:%s", config.MongoHost, config.MongoPort))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	if config.MongoPass != "" {
		err = database.session.Login(&mgo.Credential{
			Source:   config.MongoName,
			Username: config.MongoUser,
			Password: config.MongoPass,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to authenticate to database")
		}
	}

	err = database.ensureUserCollection(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ensure user collection")
	}
	err = database.ensureObjectCollection(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ensure object collection")
	}
	err = database.ensureRatingCollection(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ensure ratings collection")
	}
	err = database.ensureCommentCollection(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ensure ratings collection")
	}

	database.store, err = minio.New(
		fmt.Sprintf("%s:%s", config.StoreHost, config.StorePort),
		config.StoreAccess,
		config.StoreSecret,
		config.StoreSecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to object store")
	}

	exists, err := database.store.BucketExists(config.StoreBucket)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check if bucket exists")
	}

	if !exists {
		err = database.store.MakeBucket(config.StoreBucket, config.StoreLocation)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create bucket")
		}
	}

	exists, err = database.store.BucketExists(config.StoreBucket)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check it bucket exists after creation")
	}
	if !exists {
		return nil, errors.New("bucket was not created")
	}

	database.StoreBucket = config.StoreBucket
	database.StoreLocation = config.StoreLocation

	return &database, nil
}

// CollectionExists checks if a collection exists in MongoDB
func (database Database) CollectionExists(db, wantCollection string) (bool, error) {
	collections, err := database.session.DB(db).CollectionNames()
	if err != nil {
		return false, err
	}

	for _, collection := range collections {
		if collection == wantCollection {
			return true, nil
		}
	}

	return false, nil
}

func (database *Database) ensureUserCollection(config Config) (err error) {
	exists, err := database.CollectionExists(config.MongoName, "users")
	if err != nil {
		return err
	}
	if !exists {
		err = database.session.DB(config.MongoName).C("users").Create(&config.MongoCollectionInfo)
		if err != nil {
			return err
		}
	}
	database.users = database.session.DB(config.MongoName).C("users")

	err = database.users.EnsureIndex(mgo.Index{
		Name:   "UNIQUE_ID",
		Key:    []string{"id"},
		Unique: true,
	})
	if err != nil {
		return err
	}
	err = database.users.EnsureIndex(mgo.Index{
		Name:   "UNIQUE_NAME",
		Key:    []string{"name"},
		Unique: true,
	})
	if err != nil {
		return err
	}
	err = database.users.EnsureIndex(mgo.Index{
		Name:   "UNIQUE_EMAIL",
		Key:    []string{"email"},
		Unique: true,
	})

	return
}

func (database *Database) ensureObjectCollection(config Config) (err error) {
	exists, err := database.CollectionExists(config.MongoName, "objects")
	if err != nil {
		return err
	}
	if !exists {
		err = database.session.DB(config.MongoName).C("objects").Create(&config.MongoCollectionInfo)
		if err != nil {
			return err
		}
	}
	database.objects = database.session.DB(config.MongoName).C("objects")

	err = database.objects.EnsureIndex(mgo.Index{
		Name:   "UNIQUE_OBJECT_ID",
		Key:    []string{"id"},
		Unique: true,
	})

	return
}

func (database *Database) ensureRatingCollection(config Config) (err error) {
	exists, err := database.CollectionExists(config.MongoName, "ratings")
	if err != nil {
		return err
	}
	if !exists {
		err = database.session.DB(config.MongoName).C("ratings").Create(&config.MongoCollectionInfo)
		if err != nil {
			return err
		}
	}
	database.ratings = database.session.DB(config.MongoName).C("ratings")

	err = database.ratings.EnsureIndex(mgo.Index{
		Name:   "UNIQUE_USER_OBJECT_RATING",
		Key:    []string{"userid", "objectid"},
		Unique: true,
	})

	return
}

func (database *Database) ensureCommentCollection(config Config) (err error) {
	exists, err := database.CollectionExists(config.MongoName, "comments")
	if err != nil {
		return err
	}
	if !exists {
		err = database.session.DB(config.MongoName).C("comments").Create(&config.MongoCollectionInfo)
		if err != nil {
			return err
		}
	}
	database.comments = database.session.DB(config.MongoName).C("comments")

	return
}
