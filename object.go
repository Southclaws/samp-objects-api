package main

import (
	"net/http"

	"github.com/gorilla/mux"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

// Objects handles the /objects/:object endpoint, it includes actions for listing all objects,
// viewing info for a specific object and uploading a new object.
func (app App) Objects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if len(vars) == 0 {
		app.listObjects(w, r)
		return
	}

	objectID := types.ObjectID(vars["objectid"])
	if err := objectID.Validate(); err != nil {
		// todo: error
		return
	}

	//
}

// list objects
func (app App) listObjects(w http.ResponseWriter, r *http.Request) {

}

// get object
func (app App) getObject(w http.ResponseWriter, r *http.Request) {

}

// post object
func (app App) createObject(w http.ResponseWriter, r *http.Request) {

}
