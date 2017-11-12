package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"bitbucket.org/Southclaws/samp-objects-api/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// UploadResponse is the response returned when a client successfully uploads a file
// this is for FineUploader
type UploadResponse struct {
	Success bool `json:"success"`
}

// Objects handles the /objects/:object endpoint, it includes actions for listing all objects,
// viewing info for a specific object and uploading a new object.
func (app App) Objects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if len(vars) == 0 {

		// No object ID specified:
		// Either get all objects or propose a new object and get a URL to upload to
		switch r.Method {
		case "GET":
			app.listObjects(w, r)
		case "POST":
			app.prepObject(w, r)
		}

	} else {

		// Object ID specified:
		// Either get an object's info or update an existing object
		switch r.Method {
		case "GET":
			app.getObject(w, r)
		case "POST":
			app.updateObject(w, r)
		}

	}

	//
}

// list objects
func (app App) listObjects(w http.ResponseWriter, r *http.Request) {

}

// post object
func (app App) prepObject(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	var object types.Object
	err = json.Unmarshal(payload, &object)
	if err != nil {
		return
	}

	err = object.ValidatePartial()
	if err != nil && err.Error() != "id does not match pattern" {
		return
	}

	object.ID = types.ObjectID(uuid.New().String())

	app.Pending.Set(string(object.ID), object.ID, time.Minute*30)

	payload, err = json.Marshal(object)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(payload)
}

// ObjectUpload is the file upload endpoint
func (app App) ObjectUpload(w http.ResponseWriter, r *http.Request) {
	resp := &UploadResponse{}
	defer func() {
		payload, err := json.Marshal(resp)
		if err != nil {
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(payload))
	}()

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	fmt.Println(len(payload))
	fmt.Println(payload)

}

// get object
func (app App) getObject(w http.ResponseWriter, r *http.Request) {
	// objectID := types.ObjectID(vars["objectid"])
	// if err := objectID.Validate(); err != nil {
	// 	// todo: error
	// 	return
	// }
	w.WriteHeader(http.StatusNotImplemented)
}

// update object
func (app App) updateObject(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
