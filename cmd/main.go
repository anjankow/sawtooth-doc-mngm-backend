package main

import (
	"doc-management/internal/app"
	"doc-management/internal/config"
	"doc-management/internal/hashing"
	"doc-management/internal/ports/http"
	"doc-management/internal/repository/mongodb"
	"log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger, err := getLogger()
	if err != nil {
		log.Fatalln("setting up the logger failed: ", err)
		return
	}
	defer logger.Sync()

	logger.Info("application started")

	hashing.Initialize(logger)
	db, err := mongodb.NewConnection(logger, config.GetDbConnectionURI())
	if err != nil {
		logger.Fatal("failed to connect to the db: " + err.Error())
	}
	defer db.Disconnect()

	app := app.NewApp(logger, db)
	if err := app.Start(); err != nil {
		logger.Fatal("failed to start the app: " + err.Error())
	}
	ser := http.NewServer(logger, &app, config.GetPort())
	if err := ser.Run(); err != nil {
		logger.Fatal("failed to run the server: " + err.Error())
	}

	logger.Info("application finished")
}

func getLogger() (*zap.Logger, error) {
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zap.FatalLevel),
	}

	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	config.Development = true
	config.Level.SetLevel(zap.DebugLevel)

	logger, err := config.Build()
	return logger.WithOptions(options...), err
}
