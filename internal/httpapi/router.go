package httpapi

import (
	"encoding/json"
	"net/http"

	"cloudpulse/internal/incident"
	"cloudpulse/internal/service"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	serviceHandler *service.Handler,
	incidentHandler *incident.Handler,
) http.Handler {
	router := chi.NewRouter()

	router.Get(
		"/health",
		healthHandler,
	)

	router.Route(
		"/api/services",
		func(router chi.Router) {
			router.Post(
				"/",
				serviceHandler.Create,
			)

			router.Get(
				"/",
				serviceHandler.FindAll,
			)

			router.Get(
				"/{id}",
				serviceHandler.FindByID,
			)

			router.Delete(
				"/{id}",
				serviceHandler.Delete,
			)
		},
	)

	router.Route(
		"/api/incidents",
		func(router chi.Router) {
			router.Get(
				"/",
				incidentHandler.FindAll,
			)

			router.Get(
				"/{id}",
				incidentHandler.FindByID,
			)
		},
	)

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

	writer.WriteHeader(
		http.StatusOK,
	)

	_ = json.NewEncoder(
		writer,
	).Encode(
		map[string]string{
			"status": "ok",
		},
	)
}
