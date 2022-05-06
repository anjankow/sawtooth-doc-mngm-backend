package main

import (
	"doc-management/internal/app"
	"doc-management/internal/hashing"
	"doc-management/internal/ports/http"
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

	app := &app.App{}
	ser := http.NewServer(logger, app, ":8077")
	if err := ser.Run(); err != nil {
		logger.Error("failed to run the server: " + err.Error())
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
