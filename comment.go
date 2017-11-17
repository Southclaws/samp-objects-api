package main

import (
	"log"
	"net/http"

	"github.com/pkg/errors"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

// CommentList handles the GET /comments/{objectid} endpoint and returns a list of comments for the
// specified object ID
func (app *App) CommentList(w http.ResponseWriter, r *http.Request) {
	WriteResponseError(w, http.StatusNotImplemented, nil)
}

// CommentCreate handles the POST /comments/{objectid} endpoint and creates a comment on the
// specified object from the requesting user
func (app *App) CommentCreate(w http.ResponseWriter, r *http.Request) {
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

	log.Println(userID)
}

// CommentUpdate handles the PATCH /comments/{commentid} endpoint
func (app *App) CommentUpdate(w http.ResponseWriter, r *http.Request) {
	//
}

// CommentRemove handles the DELETE /comments/{commentid} endpoint
func (app *App) CommentRemove(w http.ResponseWriter, r *http.Request) {
	//
}
