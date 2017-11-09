package main

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// Route defines an API call route and links it with a function call
type Route struct {
	Name          string   `json:"name"`
	Methods       []string `json:"method"`
	Path          string   `json:"path"`
	Authenticated bool     `json:"authenticated"`
	handler       http.HandlerFunc
}

func (app App) routes() (routes []Route) {
	index := func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		err := enc.Encode(routes)
		if err != nil {
			logger.Fatal("failed to write index response", zap.Error(err))
		}
	}
	routes = []Route{
		{
			Name:          "index",
			Methods:       []string{"OPTIONS", "GET"},
			Path:          "/v0/index",
			Authenticated: false,
			handler:       index,
		},
		// /accounts/
		{
			Name:          "login",
			Methods:       []string{"OPTIONS", "POST"},
			Path:          "/v0/accounts/login",
			Authenticated: false,
			handler:       app.Login,
		},
		{
			Name:          "register",
			Methods:       []string{"OPTIONS", "POST"},
			Path:          "/v0/accounts/register",
			Authenticated: false,
			handler:       app.Register,
		},
		{
			Name:          "info",
			Methods:       []string{"OPTIONS", "GET"},
			Path:          "/v0/accounts/info",
			Authenticated: true,
			handler:       app.Info,
		},
		// /objects/
		{
			Name:          "list",
			Methods:       []string{"OPTIONS", "GET"},
			Path:          "/v0/objects/:objectid",
			Authenticated: true,
			handler:       app.Objects,
		},
	}
	return
}
