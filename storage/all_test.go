package storage

import (
	"os"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

var db *Database

func TestMain(m *testing.M) {
	var err error
	db, err = New(Config{
		MongoHost:   "localhost",
		MongoPort:   "27017",
		MongoUser:   "sampobjects",
		MongoPass:   "",
		MongoName:   "sampobjects",
		StoreHost:   "localhost",
		StorePort:   "9000",
		StoreAccess: "default",
		StoreSecret: "12345678",
		StoreSecure: false,
		StoreBucket: "samp-objects",
	})
	if err != nil {
		panic(err)
	}

	if os.Getenv("NO_CLEAN") == "" {
		// clean db before tests
		_, err = db.users.RemoveAll(bson.M{})
		if err != nil {
			panic(err)
		}
		_, err = db.objects.RemoveAll(bson.M{})
		if err != nil {
			panic(err)
		}
		_, err = db.ratings.RemoveAll(bson.M{})
		if err != nil {
			panic(err)
		}

		// clean S3 bucket
		doneCh := make(chan struct{})
		infoCh := db.store.ListObjects(db.StoreBucket, "", true, doneCh)
		for info := range infoCh {
			err = db.store.RemoveObject(db.StoreBucket, info.Key)
			if err != nil {
				panic(err)
			}
		}
	}

	os.Exit(m.Run())
}
