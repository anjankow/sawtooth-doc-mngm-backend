package config

import (
	"os"
	"time"
)

const (
	DefaultDbPort = ":27017"

	defaultLocalPort      = ":8080"
	defaultDatabaseName   = "documents"
	defaultDbURI          = "mongodb://root:example@localhost:27017/"
	defaultRestAPIAddr    = "localhost:8008"
	defaultValidatorAddr  = "localhost:4004"
	defaultRequestTimeout = 10 * time.Second
)

var (
	port          string
	connectionURI string
	dbName        string
	validatorAddr string
	restAPIAddr   string
)

func GetValidatorAddr() string {
	if validatorAddr == "" {
		addr := os.Getenv("VALIDATOR_ADDR")
		if addr != "" {
			validatorAddr = addr

		} else {
			validatorAddr = defaultValidatorAddr
		}
	}

	return validatorAddr
}

func GetValidatorRestAPIAddr() string {
	if restAPIAddr == "" {
		addr := os.Getenv("VALIDATOR_RESTAPI_ADDR")
		if addr != "" {
			restAPIAddr = addr

		} else {
			restAPIAddr = defaultRestAPIAddr
		}
	}

	return restAPIAddr
}

// GetPort returns port prepended with `:`
func GetPort() string {
	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = defaultLocalPort
		} else {
			port = ":" + port
		}

	}

	return port
}

func GetDbConnectionURI() string {
	if connectionURI == "" {
		connectionURI = os.Getenv("DB_URI")
		if connectionURI == "" {

			connectionURI = defaultDbURI
		}
	}

	return connectionURI
}

func GetDatabaseName() string {
	if dbName != "" {
		return dbName
	}

	dbNameEnv := os.Getenv("DB_NAME")
	if dbNameEnv != "" {
		dbName = dbNameEnv
		return dbName
	}

	dbName = defaultDatabaseName
	return dbName
}

func GetRequestTimeout() time.Duration {
	return defaultRequestTimeout
}
