package main

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// just a collection of small utility/helper functions

// Response represents a message sent back after each request
type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// WriteResponse is for quickly writing back a response with a 200-range status and a message
func WriteResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if !bodyAllowedForStatus(status) {
		return
	}

	marshalResponse(w, Response{Message: message})
}

// WriteResponseError is for quickly writing back an error response payload to a client
func WriteResponseError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if !bodyAllowedForStatus(status) {
		logger.Error("request error (not sent to client)",
			zap.String("status", http.StatusText(status)),
			zap.Error(err))
		return
	}

	logger.Error("request error (sent to client)",
		zap.String("status", http.StatusText(status)),
		zap.Error(err))

	marshalResponse(w, Response{Error: err.Error()})
}

func marshalResponse(w http.ResponseWriter, response Response) {
	payload, err := json.Marshal(response)
	if err != nil {
		logger.Fatal("failed to marshal client response",
			zap.Error(err))
	}

	_, err = w.Write(payload)
	if err != nil {
		logger.Fatal("failed to write client response",
			zap.Error(err),
			zap.Any("response", response))
	}
}

// bodyAllowedForStatus reports whether a given response status code
// permits a body. See RFC 2616, section 4.4.
// copied from Go: net/http/transfer.go
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}
