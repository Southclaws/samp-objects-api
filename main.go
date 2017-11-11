package main

import (
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var version = "master"

// Config stores app global configuration
type Config struct {
	Version       string
	Bind          string
	Domain        string
	MongoHost     string
	MongoPort     string
	MongoName     string
	MongoUser     string
	MongoPass     string
	AuthSecret    string
	StoreHost     string
	StorePort     string
	StoreAccess   string
	StoreSecret   string
	StoreSecure   bool
	StoreBucket   string
	StoreLocation string
}

var logger *zap.Logger

func init() {
	var config zap.Config
	debug := os.Getenv("DEBUG")

	if os.Getenv("TESTING") != "" {
		config = zap.NewDevelopmentConfig()
		config.DisableStacktrace = true
		config.DisableCaller = true
	} else {
		config = zap.NewProductionConfig()
		config.DisableStacktrace = true
		config.EncoderConfig.MessageKey = "@message"
		config.EncoderConfig.TimeKey = "@timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		if debug != "0" && debug != "" {
			dyn := zap.NewAtomicLevel()
			dyn.SetLevel(zap.DebugLevel)
			config.Level = dyn
		}
	}
	_logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	logger = _logger.With(
		zap.String("@version", os.Getenv("GIT_HASH")),
		zap.Namespace("@fields"),
	)
}

func main() {
	config := Config{
		Version:       version,
		Bind:          configStrFromEnv("BIND"),
		Domain:        configStrFromEnv("DOMAIN"),
		MongoHost:     configStrFromEnv("MONGO_HOST"),
		MongoPort:     configStrFromEnv("MONGO_PORT"),
		MongoName:     configStrFromEnv("MONGO_NAME"),
		MongoUser:     configStrFromEnv("MONGO_USER"),
		MongoPass:     os.Getenv("MONGO_PASS"),
		AuthSecret:    configStrFromEnv("AUTH_SECRET"),
		StoreHost:     configStrFromEnv("STORE_HOST"),
		StorePort:     configStrFromEnv("STORE_PORT"),
		StoreAccess:   configStrFromEnv("STORE_ACCESS"),
		StoreSecret:   configStrFromEnv("STORE_SECRET"),
		StoreSecure:   configStrFromEnv("STORE_SECURE") == "true",
		StoreBucket:   configStrFromEnv("STORE_BUCKET"),
		StoreLocation: configStrFromEnv("STORE_LOCATION"),
	}
	app := Initialise(config)
	app.Start()
}

func configStrFromEnv(name string) (value string) {
	value = os.Getenv(name)
	if value == "" {
		logger.Fatal("environment variable not set",
			zap.String("name", name))
	}
	return
}

func configIntFromEnv(name string) (value int) {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		logger.Fatal("environment variable not set",
			zap.String("name", name))
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		logger.Fatal("failed to convert environment variable to int",
			zap.Error(err),
			zap.String("name", name))
	}
	return
}
