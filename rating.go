package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"bitbucket.org/Southclaws/samp-objects-api/types"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// RatingCreate handles the POST /ratings/{objectid} endpoint and attempts to add a rating to an
// object from a user, if that rating already exists, it is removed.
func (app *App) RatingCreate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objectID := types.ObjectID(vars["objectid"])

	session, err := app.Sessions.Get(r, UserSessionCookie)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to read or create cookie session, clear cookies and log in again"))
		return
	}

	userIDraw, ok := session.Values["UserID"]
	if !ok {
		WriteResponseError(w, http.StatusBadRequest, errors.Wrap(err, "failed to read user ID from session"))
		return
	}

	userID, ok := userIDraw.(types.UserID)
	if !ok {
		WriteResponseError(w, http.StatusBadRequest, errors.Wrap(err, "failed to interpret user ID as string"))
		return
	}

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
	} else {
		WriteResponse(w, http.StatusCreated, "rating created")
	}
}

// RatingList handles the GET /ratings/{objectid} endpoint and returns a list of ratings
func (app *App) RatingList(w http.ResponseWriter, r *http.Request) {
	WriteResponseError(w, http.StatusNotImplemented, nil)
}
