package http

import (
	"doc-management/internal/app"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

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

	router.HandleFunc("/api/proposals/{userID}", ser.getProposals).Methods(http.MethodGet)
	router.HandleFunc("/api/proposals/{userID}", ser.putProposal).Methods(http.MethodPut)

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

	c := cors.New(cors.Options{
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut},
		AllowCredentials: true,
		Debug:            false,
	})
	handler := c.Handler(router)
	ser.httpServer = &http.Server{
		Handler: handler,
		Addr:    ser.addr,
	}

	return ser.httpServer.ListenAndServe()
}
