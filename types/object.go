package types

import "errors"

// Object represents an object that a user has uploaded, it includes a hash of the file contents
// and details such as name and owner.
type Object struct {
	ID          string `json:"id"`
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Hash        string `json:"hash"`
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
