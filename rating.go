package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"bitbucket.org/Southclaws/samp-objects-api/types"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Ratings handles the /ratings/{userid}/{objectid}
func (app *App) Ratings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := types.UserID(vars["userid"])
	objectID := types.ObjectID(vars["objectid"])

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to read payload"))
		return
	}
	r.Body.Close()

	rating := types.Rating{}
	err = json.Unmarshal(payload, &rating)
	if err != nil {
		WriteResponseError(w, http.StatusBadRequest, errors.Wrap(err, "failed to decode payload"))
		return
	}

	exists, err := app.Storage.AddRating(userID, objectID, rating.Value)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to add rating"))
		return
	}

	if exists {
		err = app.Storage.RemoveRating(userID, objectID)
		if err != nil {
			WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to remove rating"))
			return
		}
		WriteResponse(w, http.StatusAccepted, "rating removed")
	}

	WriteResponse(w, http.StatusCreated, "rating created")
}
