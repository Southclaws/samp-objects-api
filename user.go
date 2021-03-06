package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Southclaws/samp-objects-api/types"
)

// User endpoints differ from Account endpoints as they deal with public information only such as
// profiles, ratings, statistics, etc.

// UserProfile handles the /user/:userid endpoint and returns a public User object
func (app *App) UserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userName := types.UserName(vars["username"])

	if err := userName.Validate(); err != nil {
		WriteResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	user, exists, err := app.Storage.GetUserByName(userName)
	if err != nil {
		return
	}
	if !exists {
		WriteResponse(w, http.StatusNotFound, "user does not exist")
		return
	}

	user.Password = ""

	payload, err := json.Marshal(user)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(payload)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, err)
		return
	}
}
