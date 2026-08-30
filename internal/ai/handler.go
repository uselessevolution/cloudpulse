package ai

import (
	"cloudpulse/internal/incident"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	contextBuilder *ContextBuilder
	analyzer       IncidentAnalyzer
}

func NewHandler(
	contextBuilder *ContextBuilder,
	analyzer IncidentAnalyzer,
) *Handler {
	return &Handler{
		contextBuilder: contextBuilder,
		analyzer:       analyzer,
	}
}

func (handler *Handler) AnalyzeIncident(
	writer http.ResponseWriter,
	request *http.Request,
) {
	rawID :=
		chi.URLParam(
			request,
			"id",
		)

	incidentID, err :=
		strconv.ParseInt(
			rawID,
			10,
			64,
		)

	if err != nil || incidentID <= 0 {
		writeJSON(
			writer,
			http.StatusBadRequest,
			errorResponse{
				Message: "invalid incident id",
			},
		)
		return
	}

	incidentContext, err :=
		handler.contextBuilder.Build(
			request.Context(),
			incidentID,
		)

	if errors.Is(
		err,
		incident.ErrNotFound,
	) {
		writeJSON(
			writer,
			http.StatusNotFound,
			errorResponse{
				Message: "incident not found",
			},
		)
		return
	}

	if err != nil {
		writeJSON(
			writer,
			http.StatusInternalServerError,
			errorResponse{
				Message: "failed to build incident context",
			},
		)
		return
	}

	analysis, err :=
		handler.analyzer.Analyze(
			request.Context(),
			incidentContext,
		)

	if err != nil {
		log.Printf(
			"event=ai_incident_analysis_failed incident_id=%d error=%q",
			incidentID,
			err,
		)

		writeJSON(
			writer,
			http.StatusBadGateway,
			errorResponse{
				Message: "AI incident analysis failed",
			},
		)
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		analysis,
	)
}

type errorResponse struct {
	Message string `json:"message"`
}

func writeJSON(
	writer http.ResponseWriter,
	status int,
	value any,
) {
	writer.Header().Set(
		"Content-Type",
		"application/json",
	)

	writer.WriteHeader(
		status,
	)

	_ = json.NewEncoder(
		writer,
	).Encode(
		value,
	)
}
