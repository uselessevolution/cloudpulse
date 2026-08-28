package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	router := chi.NewRouter()

	router.Get("/health", healthHandler)

	return router
}

func healthHandler(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writer.Header().Set(
		"Content-Type",
		"application/json",
	)

	writer.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(writer).Encode(
		map[string]string{
			"status": "ok",
		},
	)
}
