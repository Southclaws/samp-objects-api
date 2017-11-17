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
			Methods:       []string{"GET"},
			Path:          "/v0/index",
			Authenticated: false,
			handler:       index,
		},
		// /accounts/
		{
			Name:          "login",
			Methods:       []string{"POST"},
			Path:          "/v0/accounts/login",
			Authenticated: false,
			handler:       app.Login,
		},
		{
			Name:          "register",
			Methods:       []string{"POST"},
			Path:          "/v0/accounts/register",
			Authenticated: false,
			handler:       app.Register,
		},
		{
			Name:          "get user info",
			Methods:       []string{"GET"},
			Path:          "/v0/accounts/info",
			Authenticated: true,
			handler:       app.AccountGetInfo,
		},
		{
			Name:          "update user info",
			Methods:       []string{"PATCH"},
			Path:          "/v0/accounts/info",
			Authenticated: true,
			handler:       app.AccountUpdateInfo,
		},
		// /objects/
		{
			Name:          "list objects",
			Methods:       []string{"GET"},
			Path:          "/v0/objects",
			Authenticated: false,
			handler:       app.ObjectsList,
		},
		{
			Name:          "get object info",
			Methods:       []string{"GET"},
			Path:          "/v0/objects/{userName}/{objectName}",
			Authenticated: false,
			handler:       app.ObjectByName,
		},
		// /images/
		{
			Name:          "get object image",
			Methods:       []string{"GET"},
			Path:          "/v0/images/{objectid}",
			Authenticated: false,
			handler:       app.ObjectThumb,
		},
		{
			Name:          "get object image by name",
			Methods:       []string{"GET"},
			Path:          "/v0/images/{objectid}/{fileName}",
			Authenticated: false,
			handler:       app.ObjectFiles,
		},
		// /files/
		{
			Name:          "get object file by name",
			Methods:       []string{"GET"},
			Path:          "/v0/files/{objectid}/{fileName}",
			Authenticated: false,
			handler:       app.ObjectFiles,
		},
		// /object/
		{
			Name:          "prepare object upload",
			Methods:       []string{"POST"},
			Path:          "/v0/object/prepare",
			Authenticated: true,
			handler:       app.ObjectPrepare,
		},
		{
			Name:          "upload object files",
			Methods:       []string{"POST"},
			Path:          "/v0/object/upload/{objectid}",
			Authenticated: true,
			handler:       app.ObjectUpload,
		},
		{
			Name:          "finish object upload",
			Methods:       []string{"POST"},
			Path:          "/v0/object/finish/{objectid}",
			Authenticated: true,
			handler:       app.ObjectFinish,
		},
		// /users/
		{
			Name:          "get user public profile",
			Methods:       []string{"GET"},
			Path:          "/v0/users/{username}",
			Authenticated: false,
			handler:       app.UserProfile,
		},
		// /ratings/
		{
			Name:          "post rating to object",
			Methods:       []string{"POST"},
			Path:          "/v0/ratings/{objectid}",
			Authenticated: true,
			handler:       app.RatingCreate,
		},
		{
			Name:          "list object ratings",
			Methods:       []string{"DELETE"},
			Path:          "/v0/ratings/{objectid}",
			Authenticated: true,
			handler:       app.RatingList,
		},
		// /comments/
		{
			Name:          "list comments",
			Methods:       []string{"GET"},
			Path:          "/v0/comments/{objectid}",
			Authenticated: true,
			handler:       app.CommentList,
		},
		{
			Name:          "post comment to object",
			Methods:       []string{"GET"},
			Path:          "/v0/comments/{objectid}",
			Authenticated: true,
			handler:       app.CommentCreate,
		},
		{
			Name:          "update comment",
			Methods:       []string{"PATCH"},
			Path:          "/v0/comments/{objectid}/{commentid}",
			Authenticated: true,
			handler:       app.CommentUpdate,
		},
		{
			Name:          "remove comment",
			Methods:       []string{"PATCH"},
			Path:          "/v0/comments/{objectid}/{commentid}",
			Authenticated: true,
			handler:       app.CommentRemove,
		},
	}
	return
}
