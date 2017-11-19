package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/Southclaws/samp-objects-api/storage"
	"github.com/Southclaws/samp-objects-api/types"
)

// Register handles creating an account for a user
func (app App) Register(w http.ResponseWriter, r *http.Request) {
	user := types.User{}

	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to read request body"))
		return
	}
	r.Body.Close()

	session, err := app.Sessions.Get(r, UserSessionCookie)
	if err != nil && err.Error() != "securecookie: the value is not valid" {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to read or create cookie session"))
		return
	}

	err = json.Unmarshal(raw, &user)
	if err != nil {
		WriteResponseError(w, http.StatusBadRequest, errors.Wrap(err, "failed to decode request body"))
		return
	}

	// ensure password field is a SHA256
	if len(user.Password) != 64 {
		WriteResponseError(w, http.StatusBadRequest, errors.Wrap(err, "password not in valid format"))
		return
	}

	// bcrypt the user.Password and store it back to the user.Password field
	passhash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Fatal("failed to generate bcrypt from sha256")
	}
	user.Password = types.UserPass(passhash)

	// generate a unique ID for the user
	user.ID = types.UserID(uuid.New().String())

	err = app.Storage.CreateUser(user)
	if err != nil {
		if err == storage.ErrUserNameAlreadyExists {
			WriteResponse(w, http.StatusConflict, "username already registered")
		} else if err == storage.ErrUserEmailAlreadyExists {
			WriteResponse(w, http.StatusTeapot, "email already registered")
		} else {
			WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to create new user"))
		}
		return
	}

	session.Values["UserID"] = user.ID

	app.WriteToken(w, r, session, user.ID)
}

// Login handles authentication, returns a JWT token on success
func (app App) Login(w http.ResponseWriter, r *http.Request) {
	var authRequest AuthRequest
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to read request body"))
		return
	}
	r.Body.Close()

	session, err := app.Sessions.Get(r, UserSessionCookie)
	if err != nil && err.Error() != "securecookie: the value is not valid" {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to read or create cookie session"))
		return
	}

	err = json.Unmarshal(raw, &authRequest)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to decode request body"))
		return
	}

	user, exists, err := app.Storage.GetUserByName(authRequest.Username)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to lookup user by name"))
		return
	}
	if !exists {
		WriteResponse(w, http.StatusUnauthorized, "user not found")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(authRequest.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			WriteResponseError(w, http.StatusUnauthorized, errors.New("invalid password"))
		} else {
			WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to process password"))
		}
		return
	}

	session.Values["UserID"] = user.ID

	app.WriteToken(w, r, session, user.ID)
}

// Refresh refreshes a JWT token
func (app App) Refresh(w http.ResponseWriter, r *http.Request) {
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

	app.WriteToken(w, r, session, userID)
}

// AccountGetInfo returns a types.User object for the user making the request
func (app App) AccountGetInfo(w http.ResponseWriter, r *http.Request) {
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

	user, exists, err := app.Storage.GetUser(userID)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to get user object"))
		return
	}
	if !exists {
		WriteResponse(w, http.StatusNotFound, "user not found")
		return
	}

	// blank out the password hash, this request does not need it
	user.Password = ""

	payload, err := json.Marshal(user)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to encode response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(payload)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to write payload"))
		return
	}
}

// AccountUpdateInfo updates a user's information
func (app *App) AccountUpdateInfo(w http.ResponseWriter, r *http.Request) {
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

	user := types.User{}
	err = json.Unmarshal(payload, &user)
	if err != nil {
		WriteResponseError(w, http.StatusBadRequest, errors.Wrap(err, "failed to decode payload"))
		return
	}

	if user.ID != userID {
		WriteResponseError(w, http.StatusBadRequest, errors.Wrap(err, "request user ID does not match request payload"))
		return
	}

	err = user.Validate()
	if err != nil {
		WriteResponseError(w, http.StatusBadRequest, errors.Wrap(err, "malformed user object"))
		return
	}

	err = app.Storage.UpdateUser(user)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to update user info"))
		return
	}
}
