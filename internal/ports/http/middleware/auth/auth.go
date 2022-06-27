package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"gopkg.in/square/go-jose.v2/jwt"
)

const policyName = "B2C_1_singin"

type JwtTokenParams struct {
	Issuer   string
	Audience string
}

type TokenValidator struct {
	JwtTokenParams
	logger *zap.Logger
}

func NewTokenValidator(logger *zap.Logger, params JwtTokenParams) TokenValidator {
	return TokenValidator{logger: logger, JwtTokenParams: params}
}

func (t TokenValidator) ValidateGetScopes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		claims, err := parseToken(strings.TrimPrefix(token, "Bearer "))
		if err != nil {
			t.authError(w, errors.New("failed to parse the auth token: "+err.Error()))
			return
		}

		if err := t.validateClaims(claims); err != nil {
			t.authError(w, errors.New("auth token validation: "+err.Error()))
			return
		}

		// add user id and scopes to the request context
		newCtx := r.Context()
		if user, ok := claims["oid"]; ok {
			newCtx = context.WithValue(newCtx, "userID", user)
		}
		if scopes, ok := claims["scp"]; ok {
			newCtx = context.WithValue(newCtx, "scopes", scopes)
		}

		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func (t TokenValidator) authError(w http.ResponseWriter, err error) {
	t.logger.Warn(err.Error())
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(err.Error()))
}

// dirty solution, should be done with an approperiate library;
// validating azure tokens using their tools fails
// and azure tokens are incompatible with the standard JWT validation libraries
func (t TokenValidator) validateClaims(claims map[string]interface{}) (err error) {
	if issuer, ok := claims["iss"]; ok {
		issuerStr := issuer.(string)
		if issuerStr != t.Issuer {
			err = multierr.Append(err, errors.New("invalid issuer: "+issuerStr+", expected: "+t.Issuer))
		}
	} else {
		err = multierr.Append(err, errors.New("issuer is missing"))
	}

	if audience, ok := claims["aud"]; ok {
		audienceStr := audience.(string)
		if audienceStr != t.Audience {
			err = multierr.Append(err, errors.New("invalid audience: "+audienceStr+", expected: "+t.Audience))
		}
	} else {
		err = multierr.Append(err, errors.New("audience is missing"))
	}

	if policy, ok := claims["tfp"]; ok {
		policyStr := policy.(string)
		if policyStr != policyName {
			err = multierr.Append(err, errors.New("invalid policy name: "+policyStr))
		}
	} else {
		err = multierr.Append(err, errors.New("policy is missing"))
	}

	if exp, ok := claims["exp"]; ok {

		expFloat := exp.(float64)

		expiration := time.Unix(int64(expFloat), 0)

		if expiration.Before(time.Now()) {
			err = multierr.Append(err, errors.New("token expired, expiration time: "+expiration.String()))
		}
	} else {
		err = multierr.Append(err, errors.New("expiration time is missing"))
	}

	return err
}

func parseToken(tokenString string) (map[string]interface{}, error) {

	var claims map[string]interface{}

	token, err := jwt.ParseSigned(tokenString)
	if err != nil {
		return nil, err
	}

	if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return nil, err
	}

	return claims, nil
}
