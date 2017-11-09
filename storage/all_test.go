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
		MongoHost: "localhost",
		MongoPort: "27017",
		MongoUser: "sampobjects",
		MongoPass: "",
		MongoName: "sampobjects",
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
	}

	os.Exit(m.Run())
}
