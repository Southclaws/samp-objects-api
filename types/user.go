package types

import "errors"

// User represents a user in the system, it contains their profile details and password hash
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
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
