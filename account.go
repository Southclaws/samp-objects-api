package main

import (
	"encoding/json"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
)

// Login handles authentication, returns a JWT token on success
func (app App) Login(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		success     bool
		tokenObj    *jwt.Token
		tokenString string
		claims      jwt.MapClaims
		auth        AuthResponse
		payload     []byte
	)

	success, err = app.AuthenticateLoginRequest(r)
	if err != nil {
		logger.Error("failed to authenticate login request",
			zap.Error(err))
		return
	}

	if success {
		tokenObj = jwt.New(jwt.SigningMethodHS256)
		claims = tokenObj.Claims.(jwt.MapClaims)

		claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

		tokenString, err = tokenObj.SignedString([]byte(app.config.AuthSecret))
		if err != nil {
			logger.Error("failed to sign login tokenObj",
				zap.Error(err))
			return
		}

		logger.Debug("accepted authentication")

		auth = AuthResponse{
			Message: "success",
			Token:   tokenString,
		}
	} else {
		logger.Debug("denying authentication")
		auth = AuthResponse{
			Message: "failure",
		}
		w.WriteHeader(http.StatusUnauthorized)
	}

	payload, err = json.Marshal(auth)
	if err != nil {
		logger.Error("failed to marshall token payload",
			zap.Error(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

// Register handles creating an account for a user
func (app App) Register(w http.ResponseWriter, r *http.Request) {

}
