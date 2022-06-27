package http

import (
	"doc-management/internal/app"
	"doc-management/internal/config"
	"errors"
	"net/http"
	"strings"

	"doc-management/internal/ports/http/middleware/auth"
	"doc-management/internal/ports/http/middleware/cors"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var ErrSearchTooBroad = errors.New("query is too broad, include filters")

type server struct {
	app        *app.App
	httpServer *http.Server
	addr       string
	logger     *zap.Logger
}

func (ser server) badRequest(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusBadRequest)
	ser.logger.Warn(message)
}

func (ser server) serverError(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusInternalServerError)
	ser.logger.Error(message)
}

func (ser server) registerHandlers(router *mux.Router) {

	router.HandleFunc("/health", healthcheck)

	// to create a new proposal
	router.HandleFunc("/api/proposals/{docName}", ser.putProposal).Methods(http.MethodPut)
	// to sign a certain proposal
	router.HandleFunc("/api/proposals/{proposalID}", ser.signProposal).Methods(http.MethodPost)

	// for getting all proposals filtered by a certain category or author
	router.HandleFunc("/api/proposals", ser.getAllProposals).Methods(http.MethodGet)

	// for getting all docs by certain author, signed by a certain user, ...
	router.HandleFunc("/api/docs", ser.getDocuments).Methods(http.MethodGet)
	// for getting all versions of a certain doc
	router.HandleFunc("/api/docs/{category}/{docName}", ser.getDocVersions).Methods(http.MethodGet)

}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("all good here"))
}

func NewServer(logger *zap.Logger, a *app.App, address string) server {
	return server{
		app:    a,
		addr:   address,
		logger: logger,
	}
}

func (ser server) Run() error {
	router := mux.NewRouter()
	ser.registerHandlers(router)

	tokenValidator := auth.NewTokenValidator(ser.logger, auth.JwtTokenParams{
		Issuer:   config.GetTokenIssuer(),
		Audience: config.GetClientID(),
	})

	handler := cors.AddCorsPolicy(tokenValidator.ValidateGetScopes(router))

	ser.httpServer = &http.Server{
		Handler: handler,
		Addr:    ser.addr,
	}

	ser.logger.Info("running HTTP server on " + ser.addr)
	return ser.httpServer.ListenAndServe()
}

func normalize(str string) string {
	return strings.TrimSpace(str)
}
