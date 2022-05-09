package config

import "os"

const (
	defaultLocalPort    = ":8080"
	defaultDatabaseName = "documents"
	defaultDbURI        = "mongodb://root:example@localhost:27017/"

	DefaultDbPort = ":27017"
)

var (
	port          string
	connectionURI string
	dbName        string
)

func GetValidatorRestApiAddr() string {
	return "http://localhost:8008"
}

// GetPort returns port prepended with `:`
func GetPort() string {
	if port != "" {
		return port
	}

	portNum := os.Getenv("PORT")
	if portNum != "" {
		port = ":" + portNum
		return port
	}

	port = defaultLocalPort
	return port
}

func GetDbConnectionURI() string {
	if connectionURI != "" {
		return connectionURI
	}

	connectionURI = os.Getenv("DB_URI")
	if connectionURI != "" {
		return connectionURI
	}

	connectionURI = defaultDbURI

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
