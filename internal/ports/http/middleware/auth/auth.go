package auth

import (
	"context"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"go.uber.org/zap"
)

const issuerURL = "https://csunivie3.b2clogin.com/20127c08-9cf9-41d0-a3ee-1fd0c4e787b8/v2.0/"

var audience = []string{"bec9d628-b94f-474f-a681-0abf30268fde"}

// source: https://pkg.go.dev/github.com/auth0/go-jwt-middleware/v2#section-readme
func NewJwtMiddleware(secret string, logger *zap.Logger) (*jwtmiddleware.JWTMiddleware, error) {
	keyFunc := func(ctx context.Context) (interface{}, error) {
		// Our token must be signed using this data.
		return []byte(secret), nil
	}

	// Set up the validator.
	jwtValidator, err := validator.New(
		keyFunc,
		validator.RS256,
		issuerURL,
		audience,
	)
	if err != nil {
		return nil, err
	}

	errorHndl := func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Warn("failed to auth the request: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}
	// Set up the middleware.
	middleware := jwtmiddleware.New(jwtValidator.ValidateToken, jwtmiddleware.WithErrorHandler(errorHndl), jwtmiddleware.WithTokenExtractor(jwtmiddleware.AuthHeaderTokenExtractor))

	return middleware, nil
}

func getTokenInfo(r *http.Request) validator.RegisteredClaims {
	claims := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	return claims.RegisteredClaims
}

func AddTokenValidation(middleware *jwtmiddleware.JWTMiddleware, handler http.Handler) http.Handler {
	return middleware.CheckJWT(handler)
}
