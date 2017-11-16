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

// ObjectRateCount represents the number of ratings an object has received
type ObjectRateCount int

// ObjectRateTotal represents the total sum of all ratings on an object
type ObjectRateTotal float64

// File represents an object's content filename
type File string

// Object represents an object that a object has uploaded, it includes a hash of the file contents
// and details such as name and owner.
type Object struct {
	ID          ObjectID          `json:"id"`
	OwnerID     UserID            `json:"owner_id"`
	OwnerName   UserName          `json:"owner_name"`
	Name        ObjectName        `json:"name"`
	Description ObjectDescription `json:"description"`
	Category    ObjectCategory    `json:"category"`
	Tags        []ObjectTag       `json:"tags"`
	RateCount   ObjectRateCount   `json:"rate_count"`
	RateTotal   ObjectRateTotal   `json:"rate_value"`
	RateAverage float64           `json:"rate_average" bson:"-"` // not stored in db
	Images      []File            `json:"images"`
	Models      []File            `json:"models"`
	Textures    []File            `json:"textures"`
}

// ObjectFile represents a single file the user uploaded
type ObjectFile struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ObjectDFF represents a model file
type ObjectDFF ObjectFile

// ObjectTXD represents a texture file
type ObjectTXD ObjectFile

// ObjectFiles represents a collection of models and textures
type ObjectFiles struct {
	Models   []ObjectDFF `json:"models"`
	Textures []ObjectTXD `json:"textures"`
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
	if object.OwnerID == "" {
		return errors.New("owner id is empty")
	}
	if object.OwnerName == "" {
		return errors.New("owner name is empty")
	}
	if object.Name == "" {
		return errors.New("name is empty")
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
	if len(object.Images) == 0 {
		return errors.New("no images in object")
	}
	if len(object.Models) == 0 {
		return errors.New("no models in object")
	}
	if len(object.Textures) == 0 {
		return errors.New("no textures in object")
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
