package main

import (
	"context"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"bitbucket.org/Southclaws/samp-objects-api/storage"
)

// App stores global state for routing and coordination
type App struct {
	ctx     context.Context
	cancel  context.CancelFunc
	config  Config
	router  *mux.Router
	Storage *storage.Database
}

// Initialise sets up a database connection, binds all the routes and prepares for Start
func Initialise(config Config) *App {
	var err error
	logger.Debug("initialising samp-servers-api with debug logging", zap.Any("config", config))

	app := App{
		config: config,
	}
	app.ctx, app.cancel = context.WithCancel(context.Background())

	app.Storage, err = storage.New(storage.Config{
		MongoHost: config.MongoHost,
		MongoPort: config.MongoPort,
		MongoUser: config.MongoUser,
		MongoPass: config.MongoPass,
		MongoName: config.MongoName,
	})
	if err != nil {
		logger.Fatal("failed to interact with database",
			zap.Error(err))
	}
	app.SetupAuth()

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

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	err := http.ListenAndServe(app.config.Bind, handlers.CORS(headersOk, originsOk, methodsOk)(app.router))

	logger.Fatal("http server encountered fatal error",
		zap.Error(err))
}
