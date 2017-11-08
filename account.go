package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"bitbucket.org/Southclaws/samp-objects-api/storage"
	"bitbucket.org/Southclaws/samp-objects-api/types"
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
	if err != nil {
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

	passhash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Fatal("failed to generate bcrypt from sha256")
	}
	user.Password = types.UserPass(passhash)

	err = app.Storage.CreateUser(user)
	if err != nil {
		if err == storage.ErrUsernameAlreadyExists {
			WriteResponseError(w, http.StatusConflict, errors.Wrap(err, "username already registered"))
		} else {
			WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to create new user"))
		}
		return
	}

	err = app.WriteToken(w, r, session)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to write token response"))
		return
	}
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
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to read or create cookie session"))
		return
	}

	err = json.Unmarshal(raw, &authRequest)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to decode request body"))
		return
	}

	user, err := app.Storage.GetUserByName(authRequest.Username)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to lookup user by name"))
		return
	}

	logger.Debug("pw",
		zap.String("db", string(user.Password)),
		zap.String("pl", string(authRequest.Password)))

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(authRequest.Password)); err != nil {
		WriteResponseError(w, http.StatusUnauthorized, errors.Wrap(err, "invalid password"))
		return
	}

	err = app.WriteToken(w, r, session)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to write token response"))
		return
	}
}

// Refresh refreshes a JWT token
func (app App) Refresh(w http.ResponseWriter, r *http.Request) {
	session, err := app.Sessions.Get(r, UserSessionCookie)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to read or create cookie session"))
		return
	}

	err = app.WriteToken(w, r, session)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to write token response"))
		return
	}
}

// Info returns a types.User object for the user making the request
func (app App) Info(w http.ResponseWriter, r *http.Request) {

}
