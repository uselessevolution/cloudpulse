package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repository *Repository
}

func NewHandler(
	repository *Repository,
) *Handler {
	return &Handler{
		repository: repository,
	}
}

type CreateRequest struct {
	Name                 string `json:"name"`
	URL                  string `json:"url"`
	ExpectedStatus       int    `json:"expectedStatus"`
	CheckIntervalSeconds int    `json:"checkIntervalSeconds"`
	TimeoutSeconds       int    `json:"timeoutSeconds"`
	Enabled              *bool  `json:"enabled"`
}

type errorResponse struct {
	Message string `json:"message"`
}

func (handler *Handler) Create(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input CreateRequest

	if err := json.NewDecoder(
		request.Body,
	).Decode(&input); err != nil {
		writeJSON(
			writer,
			http.StatusBadRequest,
			errorResponse{
				Message: "invalid JSON body",
			},
		)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.URL = strings.TrimSpace(input.URL)

	if input.Name == "" {
		writeJSON(
			writer,
			http.StatusBadRequest,
			errorResponse{
				Message: "name is required",
			},
		)
		return
	}

	if input.URL == "" {
		writeJSON(
			writer,
			http.StatusBadRequest,
			errorResponse{
				Message: "url is required",
			},
		)
		return
	}

	expectedStatus := input.ExpectedStatus
	if expectedStatus == 0 {
		expectedStatus = 200
	}

	checkIntervalSeconds :=
		input.CheckIntervalSeconds
	if checkIntervalSeconds == 0 {
		checkIntervalSeconds = 30
	}

	timeoutSeconds := input.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = 3
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	if expectedStatus < 100 ||
		expectedStatus > 599 {
		writeJSON(
			writer,
			http.StatusBadRequest,
			errorResponse{
				Message: "expectedStatus must be between 100 and 599",
			},
		)
		return
	}

	if checkIntervalSeconds <= 0 {
		writeJSON(
			writer,
			http.StatusBadRequest,
			errorResponse{
				Message: "checkIntervalSeconds must be greater than 0",
			},
		)
		return
	}

	if timeoutSeconds <= 0 {
		writeJSON(
			writer,
			http.StatusBadRequest,
			errorResponse{
				Message: "timeoutSeconds must be greater than 0",
			},
		)
		return
	}

	created, err := handler.repository.Create(
		request.Context(),
		CreateParams{
			Name:                 input.Name,
			URL:                  input.URL,
			ExpectedStatus:       expectedStatus,
			CheckIntervalSeconds: checkIntervalSeconds,
			TimeoutSeconds:       timeoutSeconds,
			Enabled:              enabled,
		},
	)

	if err != nil {
		writeJSON(
			writer,
			http.StatusInternalServerError,
			errorResponse{
				Message: "failed to create service",
			},
		)
		return
	}

	writeJSON(
		writer,
		http.StatusCreated,
		created,
	)
}

func (handler *Handler) FindAll(
	writer http.ResponseWriter,
	request *http.Request,
) {
	services, err :=
		handler.repository.FindAll(
			request.Context(),
		)

	if err != nil {
		writeJSON(
			writer,
			http.StatusInternalServerError,
			errorResponse{
				Message: "failed to load services",
			},
		)
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		services,
	)
}

func (handler *Handler) FindByID(
	writer http.ResponseWriter,
	request *http.Request,
) {
	id, err := parseID(request)

	if err != nil {
		writeJSON(
			writer,
			http.StatusBadRequest,
			errorResponse{
				Message: "invalid service id",
			},
		)
		return
	}

	found, err :=
		handler.repository.FindByID(
			request.Context(),
			id,
		)

	if errors.Is(err, ErrNotFound) {
		writeJSON(
			writer,
			http.StatusNotFound,
			errorResponse{
				Message: "service not found",
			},
		)
		return
	}

	if err != nil {
		writeJSON(
			writer,
			http.StatusInternalServerError,
			errorResponse{
				Message: "failed to load service",
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

func (handler *Handler) Delete(
	writer http.ResponseWriter,
	request *http.Request,
) {
	id, err := parseID(request)

	if err != nil {
		writeJSON(
			writer,
			http.StatusBadRequest,
			errorResponse{
				Message: "invalid service id",
			},
		)
		return
	}

	err = handler.repository.Delete(
		request.Context(),
		id,
	)

	if errors.Is(err, ErrNotFound) {
		writeJSON(
			writer,
			http.StatusNotFound,
			errorResponse{
				Message: "service not found",
			},
		)
		return
	}

	if err != nil {
		writeJSON(
			writer,
			http.StatusInternalServerError,
			errorResponse{
				Message: "failed to delete service",
			},
		)
		return
	}

	writer.WriteHeader(
		http.StatusNoContent,
	)
}

func parseID(
	request *http.Request,
) (int64, error) {
	idText := chi.URLParam(
		request,
		"id",
	)

	return strconv.ParseInt(
		idText,
		10,
		64,
	)
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

	writer.WriteHeader(status)

	_ = json.NewEncoder(
		writer,
	).Encode(value)
}
