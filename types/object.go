package types

import "errors"
import "regexp"

// ObjectID represents an object's unique ID
type ObjectID string

// ObjectName represents an object's name
type ObjectName string

// ObjectDescription represents an object's description
type ObjectDescription string

// ObjectCategory represents the category name for an object
type ObjectCategory string

// ObjectTag represents a search tag for an object
type ObjectTag string

// Hash represents an object's content hash
type Hash string

// Object represents an object that a object has uploaded, it includes a hash of the file contents
// and details such as name and owner.
type Object struct {
	ID          ObjectID          `json:"id"`
	Owner       UserID            `json:"owner"`
	Name        ObjectName        `json:"name"`
	Description ObjectDescription `json:"description"`
	Category    ObjectCategory    `json:"category"`
	Tags        []ObjectTag       `json:"tags"`
	ImageHash   Hash              `json:"image_hash"`
	ModelHash   Hash              `json:"model_hash"`
	TextureHash Hash              `json:"texture_hash"`
}

var (
	// ObjectIDMatch is a regular expression used to validate object IDs
	ObjectIDMatch = regexp.MustCompile(`[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}`)
)

// ValidatePartial is for validating pre-upload objects
func (object Object) ValidatePartial() (err error) {
	if err = object.ID.Validate(); err != nil {
		return
	}
	if object.Owner == "" {
		return errors.New("owner is empty")
	}
	if object.Name == "" {
		return errors.New("name is empty")
	}
	if object.Description == "" {
		return errors.New("description is empty")
	}
	return
}

// Validate ensures all necessary fields are correct
func (object Object) Validate() (err error) {
	if err = object.ValidatePartial(); err != nil {
		return
	}
	if object.Category == "" {
		return errors.New("category is empty")
	}
	if object.ImageHash == "" {
		return errors.New("image hash is empty")
	}
	if object.ModelHash == "" {
		return errors.New("model hash is empty")
	}
	if object.TextureHash == "" {
		return errors.New("texture hash is empty")
	}
	return
}

// Validate checks if an object ID is valid
func (objectid ObjectID) Validate() (err error) {
	if !ObjectIDMatch.MatchString(string(objectid)) {
		err = errors.New("id does not match pattern")
	}
	return
}
