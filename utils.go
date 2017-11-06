package main

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// just a collection of small utility/helper functions

// Response represents a message sent back after each request
type Response struct {
	Errors []string               `json:"errors,omitempty"`
	Extra  map[string]interface{} `json:"extra,omitempty"`
}

// WriteResponse is for quickly writing back a response payload to a client
func WriteResponse(w http.ResponseWriter, status int, errs []error, extra map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := Response{
		Errors: make([]string, len(errs)),
		Extra:  extra,
	}
	for i, err := range errs {
		logger.Error("error while processing request",
			zap.String("status", http.StatusText(status)),
			zap.Error(errors.Cause(err)),
			zap.Any("extra", extra))
		resp.Errors[i] = err.Error()
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		logger.Fatal("failed to marshal client response",
			zap.Error(err))
	}

	_, err = w.Write(payload)
	if err != nil {
		logger.Fatal("failed to write client response",
			zap.Error(err),
			zap.Any("response", resp))
	}
}
