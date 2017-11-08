package types

import "errors"

// ObjectID represents an object's unique ID
type ObjectID string

// ObjectName represents an object's name
type ObjectName string

// ObjectDescription represents an object's description
type ObjectDescription string

// ObjectHash represents an object's content hash
type ObjectHash string

// Object represents an object that a user has uploaded, it includes a hash of the file contents
// and details such as name and owner.
type Object struct {
	ID          ObjectID          `json:"id"`
	Owner       UserID            `json:"owner"`
	Name        ObjectName        `json:"name"`
	Description ObjectDescription `json:"description"`
	Hash        ObjectHash        `json:"hash"`
}

// Validate ensures all necessary fields are correct
func (user Object) Validate() (err error) {
	if user.ID == "" {
		return errors.New("id is empty")
	}
	if user.Owner == "" {
		return errors.New("owner is empty")
	}
	if user.Name == "" {
		return errors.New("name is empty")
	}
	if user.Description == "" {
		return errors.New("description is empty")
	}
	if user.Hash == "" {
		return errors.New("hash is empty")
	}
	return
}
