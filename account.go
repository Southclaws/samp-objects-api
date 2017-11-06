package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

// Register handles creating an account for a user
func (app App) Register(w http.ResponseWriter, r *http.Request) {
	user := types.User{}

	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, []error{errors.Wrap(err, "failed to read request body")}, nil)
		return
	}
	r.Body.Close()

	err = json.Unmarshal(raw, &user)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, []error{errors.Wrap(err, "failed to decode request body")}, nil)
		return
	}

	err = app.Storage.CreateUser(user)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, []error{errors.Wrap(err, "failed to create new user")}, nil)
		return
	}

	var auth AuthResponse

	tokenObj := jwt.New(jwt.SigningMethodHS256)
	claims := tokenObj.Claims.(jwt.MapClaims)

	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := tokenObj.SignedString([]byte(app.config.AuthSecret))
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, []error{errors.Wrap(err, "failed to sign authentication token")}, nil)
		return
	}

	auth = AuthResponse{
		Message: "success",
		Token:   tokenString,
	}

	payload, err := json.Marshal(auth)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, []error{errors.Wrap(err, "failed to encode authentication response")}, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

// Login handles authentication, returns a JWT token on success
func (app App) Login(w http.ResponseWriter, r *http.Request) {
	success, err := app.AuthenticateLoginRequest(r)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, []error{errors.Wrap(err, "failed to authenticate login request")}, nil)
		return
	}

	var auth AuthResponse
	if success {
		tokenObj := jwt.New(jwt.SigningMethodHS256)
		claims := tokenObj.Claims.(jwt.MapClaims)

		claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

		tokenString, err := tokenObj.SignedString([]byte(app.config.AuthSecret))
		if err != nil {
			WriteResponse(w, http.StatusInternalServerError, []error{errors.Wrap(err, "failed to sign authentication token")}, nil)
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
		WriteResponse(w, http.StatusInternalServerError, []error{errors.Wrap(err, "failed to encode authentication response")}, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}
