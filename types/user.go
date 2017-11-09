package types

import (
	"encoding/gob"
	"errors"
	"regexp"
	"strings"
)

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
	ID       UserID     `json:"id"`
	Name     UserName   `json:"name"`
	Email    UserEmail  `json:"email"`
	Password UserPass   `json:"password"`
	Objects  []ObjectID `json:"objects"`
}

var (
	// UserIDMatch is a regular expression used to validate user IDs
	UserIDMatch = regexp.MustCompile(`[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}`)

	// UserNameMatch is a regular expression used to validate user names
	// names can only contain alphanumerics and hyphens
	// other rules such as consecutive hyphens and beginning or ending with hyphens are implemented
	// by UserName.Validate()
	UserNameMatch = regexp.MustCompile(`^[a-zA-Z\d]{1,39}$`)
)

func init() {
	gob.Register(UserID(""))
}

// Validate ensures all necessary fields are correct
func (user User) Validate() (err error) {
	if user.ID == "" {
		return errors.New("id is empty")
	}
	if user.Name == "" {
		return errors.New("name is empty")
	}
	if user.Email == "" {
		return errors.New("email is empty")
	}
	if user.Password == "" {
		return errors.New("password is empty")
	}
	return
}

// Validate checks if an object ID is valid
func (userID UserID) Validate() (err error) {
	if !UserIDMatch.MatchString(string(userID)) {
		err = errors.New("id does not match pattern")
	}
	return
}

// Validate checks if an object ID is valid
func (userName UserName) Validate() (err error) {
	if len(userName) == 0 {
		return errors.New("name is null")
	}
	if len(userName) > 39 {
		return errors.New("name is over 39 characters")
	}
	if !UserNameMatch.MatchString(string(userName)) {
		return errors.New("name contains invalid characters")
	}
	if strings.HasPrefix(string(userName), "-") {
		return errors.New("name begins with a hyphen")
	}
	if strings.HasSuffix(string(userName), "-") {
		return errors.New("name ends with a hyphen")
	}
	return
}
