package main

import (
	"os"
	"strconv"

	// loads environment variables from .env
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var version = "master"

// Config stores app global configuration
type Config struct {
	Version       string
	Bind          string `split_words:"true" required:"true"`
	Domain        string `split_words:"true" required:"true"`
	MongoHost     string `split_words:"true" required:"true"`
	MongoPort     string `split_words:"true" required:"true"`
	MongoName     string `split_words:"true" required:"true"`
	MongoUser     string `split_words:"true" required:"true"`
	MongoPass     string `split_words:"true" required:"false"`
	AuthSecret    string `split_words:"true" required:"true"`
	StoreHost     string `split_words:"true" required:"true"`
	StorePort     string `split_words:"true" required:"true"`
	StoreAccess   string `split_words:"true" required:"true"`
	StoreSecret   string `split_words:"true" required:"true"`
	StoreSecure   bool   `split_words:"true" required:"true"`
	StoreBucket   string `split_words:"true" required:"true"`
	StoreLocation string `split_words:"true" required:"true"`
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
		Version: version,
	}
	err := envconfig.Process("SAMPOBJECTS", &config)
	if err != nil {
		logger.Fatal("failed to load configuration",
			zap.Error(err))
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
