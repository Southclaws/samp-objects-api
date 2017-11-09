package storage

import (
	"errors"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

var (
	// ErrUserNameAlreadyExists indicates that a user attempted to register with a name that is
	// already registered
	ErrUserNameAlreadyExists = errors.New("user already exists")

	// ErrUserEmailAlreadyExists indicates that a user attempted to register with an email that is
	// already registered
	ErrUserEmailAlreadyExists = errors.New("email already exists")
)

// CreateUser creates a new user in the database
func (db Database) CreateUser(user types.User) (err error) {
	if err = user.Validate(); err != nil {
		return
	}

	err = db.users.Insert(user)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE_NAME") {
			return ErrUserNameAlreadyExists
		} else if strings.Contains(err.Error(), "UNIQUE_EMAIL") {
			return ErrUserEmailAlreadyExists
		}
	}

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
func (db Database) DeleteUser(userID types.UserID) (err error) {
	if err = userID.Validate(); err != nil {
		return
	}

	err = db.users.Remove(bson.M{"id": userID})

	return
}

// GetUser returns a types.User by their unique ID
func (db Database) GetUser(userID types.UserID) (user types.User, err error) {
	if err = userID.Validate(); err != nil {
		return
	}

	err = db.users.Find(bson.M{"id": userID}).One(&user)
	if err != nil {
		return
	}

	return
}

// GetUserByName returns a types.User by their name
func (db Database) GetUserByName(userName types.UserName) (user types.User, err error) {
	if err = userName.Validate(); err != nil {
		return
	}

	err = db.users.Find(bson.M{"name": userName}).One(&user)
	return
}

// UserExists checks if a user exists by their unique ID
func (db Database) UserExists(userID types.UserID) (exists bool, err error) {
	if err = userID.Validate(); err != nil {
		return
	}

	count, err := db.users.Find(bson.M{"id": userID}).Count()
	if err != nil {
		return
	}
	return count != 0, err
}

// UserExistsByName checks if a user exists by their name
func (db Database) UserExistsByName(userName types.UserName) (exists bool, err error) {
	if err = userName.Validate(); err != nil {
		return
	}

	count, err := db.users.Find(bson.M{"name": userName}).Count()
	if err != nil {
		return
	}
	return count != 0, err
}
