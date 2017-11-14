package main

import (
	"context"
	"net/http"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"

	"bitbucket.org/Southclaws/samp-objects-api/storage"
	"bitbucket.org/Southclaws/samp-objects-api/types"
)

// App stores global state for routing and coordination
type App struct {
	ctx            context.Context
	cancel         context.CancelFunc
	config         Config
	router         *mux.Router
	Storage        *storage.Database
	Sessions       *sessions.CookieStore
	Uploads        sync.Map
	FinishRequests chan types.ObjectID
}

// ActiveUpload represents an object that's currently being uploaded, it contains a channel where
// new files are added and a types.Object that represents the current state of the object.
type ActiveUpload struct {
	ch     chan types.ObjectFile
	object types.Object
}

const (
	// UserSessionCookie is the name used for the Gorilla cookie storage manager
	UserSessionCookie = "userAuthData"
)

// Initialise sets up a database connection, binds all the routes and prepares for Start
func Initialise(config Config) *App {
	var err error
	logger.Debug("initialising samp-servers-api with debug logging", zap.Any("config", config))

	app := App{
		config: config,
	}
	app.ctx, app.cancel = context.WithCancel(context.Background())

	app.Storage, err = storage.New(storage.Config{
		MongoHost:     config.MongoHost,
		MongoPort:     config.MongoPort,
		MongoUser:     config.MongoUser,
		MongoPass:     config.MongoPass,
		MongoName:     config.MongoName,
		StoreHost:     config.StoreHost,
		StorePort:     config.StorePort,
		StoreAccess:   config.StoreAccess,
		StoreSecret:   config.StoreSecret,
		StoreSecure:   config.StoreSecure,
		StoreBucket:   config.StoreBucket,
		StoreLocation: config.StoreLocation,
	})
	if err != nil {
		logger.Fatal("failed to interact with database",
			zap.Error(err))
	}
	app.SetupAuth()

	// Set up session manager
	// app.Sessions = sessions.NewCookieStore(securecookie.GenerateRandomKey(64))
	app.Sessions = sessions.NewCookieStore([]byte(`securecookie.GenerateRandomKey(64)`))

	// Set up HTTP server
	app.router = mux.NewRouter().StrictSlash(true)

	for _, route := range app.routes() {
		if route.Authenticated {
			app.router.
				Methods(route.Methods...).
				Name(route.Name).
				Path(route.Path).
				Handler(app.Authenticated(route.handler))
		} else {
			app.router.
				Methods(route.Methods...).
				Name(route.Name).
				Path(route.Path).
				Handler(route.handler)
		}
	}

	return &app
}

// Start begins listening for requests and blocks until fatal error
func (app *App) Start() {
	defer app.cancel()

	err := http.ListenAndServe(app.config.Bind, handlers.CORS(
		handlers.AllowedHeaders([]string{"Cache-Control", "X-File-Name", "X-Requested-With", "X-File-Name", "Content-Type", "Authorization", "Set-Cookie", "Cookie"}),
		handlers.AllowedOrigins([]string{"https://" + app.config.Domain, "http://localhost:3000"}),
		handlers.AllowedMethods([]string{"OPTIONS", "GET", "HEAD", "POST", "PUT"}),
		handlers.AllowCredentials(),
	)(app.router))

	logger.Fatal("http server encountered fatal error",
		zap.Error(err))
}
