package healthcheck

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloudpulse/internal/service"
)

type Checker struct {
	client *http.Client
}

func NewChecker() *Checker {
	return &Checker{
		client: &http.Client{},
	}
}

func (checker *Checker) Check(
	ctx context.Context,
	target service.Service,
) Result {

	timeout := time.Duration(
		target.TimeoutSeconds,
	) * time.Second

	checkContext, cancel :=
		context.WithTimeout(
			ctx,
			timeout,
		)

	defer cancel()

	request, err :=
		http.NewRequestWithContext(
			checkContext,
			http.MethodGet,
			target.URL,
			nil,
		)

	if err != nil {
		message :=
			fmt.Sprintf(
				"create request: %v",
				err,
			)

		return Result{
			ServiceID:    target.ID,
			CheckedAt:    time.Now(),
			LatencyMS:    0,
			Success:      false,
			ErrorMessage: &message,
		}
	}

	startedAt := time.Now()

	response, err :=
		checker.client.Do(request)

	latency :=
		time.Since(startedAt)

	latencyMS :=
		latency.Milliseconds()

	if err != nil {
		message := err.Error()

		return Result{
			ServiceID:    target.ID,
			CheckedAt:    time.Now(),
			StatusCode:   nil,
			LatencyMS:    latencyMS,
			Success:      false,
			ErrorMessage: &message,
		}
	}

	defer response.Body.Close()

	_, _ = io.Copy(
		io.Discard,
		response.Body,
	)

	statusCode :=
		response.StatusCode

	success :=
		statusCode ==
			target.ExpectedStatus

	var errorMessage *string

	if !success {
		message :=
			fmt.Sprintf(
				"unexpected status code: got %d, expected %d",
				statusCode,
				target.ExpectedStatus,
			)

		errorMessage = &message
	}

	return Result{
		ServiceID:    target.ID,
		CheckedAt:    time.Now(),
		StatusCode:   &statusCode,
		LatencyMS:    latencyMS,
		Success:      success,
		ErrorMessage: errorMessage,
	}
}
