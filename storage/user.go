package storage

import (
	"errors"

	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

var (
	// ErrUsernameAlreadyExists indicates that a user attempted to register a username that already exists
	ErrUsernameAlreadyExists = errors.New("user already exists")
)

// CreateUser creates a new user in the database
func (db Database) CreateUser(user types.User) (err error) {
	exists, err := db.UserExistsByName(user.Username)
	if err != nil {
		return
	}
	if exists {
		err = ErrUsernameAlreadyExists
		return
	}

	if err = user.Validate(); err != nil {
		return
	}

	err = db.users.Insert(user)

	return
}

// UpdateUser updates a user's account information
func (db Database) UpdateUser(user types.User) (err error) {
	if err = user.Validate(); err != nil {
		return
	}

	err = db.users.Update(bson.M{"id": user.ID}, user)

	return
}

// DeleteUser deletes a user's account
func (db Database) DeleteUser(user types.User) (err error) {
	if err = user.Validate(); err != nil {
		return
	}

	err = db.users.Remove(bson.M{"id": user.ID})

	return
}

// GetUser returns a types.User by their unique ID
func (db Database) GetUser(id types.UserID) (user types.User, err error) {
	err = db.users.Find(bson.M{"id": id}).One(&user)
	if err != nil {
		return
	}

	return
}

// GetUserByName returns a types.User by their username
func (db Database) GetUserByName(username types.UserName) (user types.User, err error) {
	err = db.users.Find(bson.M{"username": username}).One(&user)
	return
}

// UserExists checks if a user exists by their unique ID
func (db Database) UserExists(id types.UserID) (exists bool, err error) {
	count, err := db.users.Find(bson.M{"id": id}).Count()
	if err != nil {
		return
	}
	return count != 0, err
}

// UserExistsByName checks if a user exists by their username
func (db Database) UserExistsByName(username types.UserName) (exists bool, err error) {
	count, err := db.users.Find(bson.M{"username": username}).Count()
	if err != nil {
		return
	}
	return count != 0, err
}
