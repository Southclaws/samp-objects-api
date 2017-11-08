package types

import "errors"
import "encoding/gob"

// UserID represents a user's unique ID
type UserID string

// UserName represents a user's name
type UserName string

// UserEmail represents a user's email address
type UserEmail string

// UserPass represents a user's password hash
type UserPass string

// User represents a user in the system, it contains their profile details and password hash
type User struct {
	ID       UserID    `json:"id"`
	Username UserName  `json:"username"`
	Email    UserEmail `json:"email"`
	Password UserPass  `json:"password"`
}

func init() {
	gob.Register(UserID(""))
}

// Validate ensures all necessary fields are correct
func (user User) Validate() (err error) {
	if user.ID == "" {
		return errors.New("id is empty")
	}
	if user.Username == "" {
		return errors.New("username is empty")
	}
	if user.Email == "" {
		return errors.New("email is empty")
	}
	if user.Password == "" {
		return errors.New("password is empty")
	}
	return
}
