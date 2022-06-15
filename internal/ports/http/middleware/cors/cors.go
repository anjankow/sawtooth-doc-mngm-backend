package cors

import (
	"net/http"

	"github.com/rs/cors"
)

func AddCorsPolicy(handler http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut},
		AllowCredentials: true,
		Debug:            false,
	})

	return c.Handler(handler)
}
