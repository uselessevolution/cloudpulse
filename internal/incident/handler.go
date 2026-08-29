package incident

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Reader interface {
	FindAll(
		ctx context.Context,
	) ([]Incident, error)

	FindByID(
		ctx context.Context,
		id int64,
	) (Incident, error)
}

type Handler struct {
	repository Reader
}

func NewHandler(
	repository Reader,
) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (handler *Handler) FindAll(
	writer http.ResponseWriter,
	request *http.Request,
) {
	incidents, err :=
		handler.repository.FindAll(
			request.Context(),
		)

	if err != nil {
		writeJSON(
			writer,
			http.StatusInternalServerError,
			errorResponse{
				Message: "failed to load incidents",
			},
		)
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		incidents,
	)
}

func (handler *Handler) FindByID(
	writer http.ResponseWriter,
	request *http.Request,
) {
	rawID :=
		chi.URLParam(
			request,
			"id",
		)

	id, err :=
		strconv.ParseInt(
			rawID,
			10,
			64,
		)

	if err != nil || id <= 0 {
		writeJSON(
			writer,
			http.StatusBadRequest,
			errorResponse{
				Message: "invalid incident id",
			},
		)
		return
	}

	found, err :=
		handler.repository.FindByID(
			request.Context(),
			id,
		)

	if errors.Is(
		err,
		ErrNotFound,
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
				Message: "failed to load incident",
			},
		)
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		found,
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
