package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

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

	err = json.Unmarshal(raw, &user)
	if err != nil {
		WriteResponseError(w, http.StatusBadRequest, errors.Wrap(err, "failed to decode request body"))
		return
	}

	err = app.Storage.CreateUser(user)
	if err != nil && err == storage.ErrUsernameAlreadyExists {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to create new user"))
		return
	}

	var auth AuthResponse

	if err == storage.ErrUsernameAlreadyExists {
		auth.Message = "username taken"
	} else {
		tokenObj := jwt.New(jwt.SigningMethodHS256)
		claims := tokenObj.Claims.(jwt.MapClaims)

		claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

		tokenString, err := tokenObj.SignedString([]byte(app.config.AuthSecret))
		if err != nil {
			WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to sign authentication token"))
			return
		}
		auth.Token = tokenString
	}

	payload, err := json.Marshal(auth)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to encode authentication response"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

// Login handles authentication, returns a JWT token on success
func (app App) Login(w http.ResponseWriter, r *http.Request) {
	success, err := app.AuthenticateLoginRequest(r)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to authenticate login request"))
		return
	}

	var auth AuthResponse
	if success {
		tokenObj := jwt.New(jwt.SigningMethodHS256)
		claims := tokenObj.Claims.(jwt.MapClaims)

		claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

		tokenString, err := tokenObj.SignedString([]byte(app.config.AuthSecret))
		if err != nil {
			WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to sign authentication token"))
			return
		}

		auth = AuthResponse{
			Message: "success",
			Token:   tokenString,
		}
	} else {
		auth = AuthResponse{
			Message: "failure",
		}
		w.WriteHeader(http.StatusUnauthorized)
	}

	payload, err := json.Marshal(auth)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to encode authentication response"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(payload)
	if err != nil {
		logger.Error("failed to write login response",
			zap.Error(err))
	}
}

// Info returns a types.User object for the user making the request
func (app App) Info(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(`{"loggedIn": "yes!"}`))
	if err != nil {
		logger.Fatal("failed to write to response writer", zap.Error(err))
	}
}
