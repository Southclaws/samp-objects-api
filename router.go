package main

import (
	"net/http"
)

// Route defines an API call route and links it with a function call
type Route struct {
	Name          string   `json:"name"`
	Methods       []string `json:"method"`
	Path          string   `json:"path"`
	Authenticated bool     `json:"authenticated"`
	handler       http.HandlerFunc
}

func (app App) routes() []Route {
	return []Route{
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
			handler:       LoggedIn,
		},
	}
}
