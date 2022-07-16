package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	DefaultDbPort = ":27017"

	defaultLocalPort      = ":8077"
	defaultDatabaseName   = "documents"
	defaultDbURI          = "mongodb://root:example@localhost:27017/"
	defaultRestAPIAddr    = "localhost:8008"
	defaultValidatorAddr  = "localhost:4004"
	defaultRequestTimeout = 10 * time.Second
)

var (
	port          string
	connectionURI string
)

func GetValidatorAddr() string {

	if addr := viper.GetString("VALIDATOR_ADDR"); addr != "" {
		return addr
	}
	return defaultValidatorAddr
}

func GetValidatorRestAPIAddr() string {

	if addr := viper.GetString("VALIDATOR_RESTAPI_ADDR"); addr != "" {
		return addr
	}
	return defaultRestAPIAddr
}

// GetPort returns port prepended with `:`
func GetPort() string {
	if port == "" {
		port = viper.GetString("PORT")
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
		connectionURI = viper.GetString("DB_URI")
		if connectionURI == "" {

			connectionURI = defaultDbURI
		}
	}

	return connectionURI
}

func GetDatabaseName() string {
	if dbNameEnv := viper.GetString("DB_NAME"); dbNameEnv != "" {
		return dbNameEnv
	}
	return defaultDatabaseName
}

func GetRequestTimeout() time.Duration {
	timeout := viper.GetDuration("REQ_TIMEOUT")
	if timeout.Seconds() < 1 {
		return defaultRequestTimeout
	}

	return timeout
}

func GetAppSecret() string {
	return viper.GetString("MS_CLIENT_SECRET")
}

func GetTenantID() string {
	return viper.GetString("MS_TENANT_ID")
}

func GetClientID() string {
	return viper.GetString("MS_CLIENT_ID")
}

func GetMsExtensionID() string {
	return viper.GetString("MS_EXTENSION_ID")
}

func GetAppUserID() string {
	return viper.GetString("APP_USER_ID")
}

func GetTokenIssuer() string {
	return viper.GetString("MS_TOKEN_ISSUER")
}
