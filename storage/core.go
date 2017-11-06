package storage

import (
	"fmt"

	"gopkg.in/mgo.v2"
)

// Database represents the storage backend state
type Database struct {
	session    *mgo.Session
	collection *mgo.Collection
}

// Config represents the configuration required to interact with the database
type Config struct {
	MongoHost           string
	MongoPort           string
	MongoUser           string
	MongoPass           string
	MongoName           string
	MongoCollection     string
	MongoCollectionInfo mgo.CollectionInfo
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
		return nil, err
	}

	if config.MongoPass != "" {
		err = database.session.Login(&mgo.Credential{
			Source:   config.MongoName,
			Username: config.MongoUser,
			Password: config.MongoPass,
		})
		if err != nil {
			return nil, err
		}
	}

	exists, err := database.CollectionExists(config.MongoName, config.MongoCollection)
	if err != nil {
		return nil, err
	}

	if !exists {
		err = database.session.DB(config.MongoName).C(config.MongoCollection).Create(&config.MongoCollectionInfo)
		if err != nil {
			return nil, err
		}
	}

	database.collection = database.session.DB(config.MongoName).C(config.MongoCollection)

	err = database.collection.EnsureIndex(mgo.Index{
		Key:    []string{"id"},
		Unique: true,
	})
	if err != nil {
		return nil, err
	}
	err = database.collection.EnsureIndex(mgo.Index{
		Key:    []string{"username"},
		Unique: true,
	})
	if err != nil {
		return nil, err
	}

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
