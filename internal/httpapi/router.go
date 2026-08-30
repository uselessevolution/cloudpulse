package httpapi

import (
	"encoding/json"
	"net/http"

	"cloudpulse/internal/incident"
	"cloudpulse/internal/service"

	"cloudpulse/internal/ai"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(
	serviceHandler *service.Handler,
	incidentHandler *incident.Handler,
	aiHandler *ai.Handler,
) http.Handler {
	router := chi.NewRouter()

	router.Get(
		"/health",
		healthHandler,
	)
	router.Handle(
		"/metrics",
		promhttp.Handler(),
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

			router.Post(
				"/{id}/ai-summary",
				aiHandler.AnalyzeIncident,
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
