package d4lerror

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

type logger interface {
	ErrInternal(context.Context, error) error
}

type ResponseErrModifier func(*ErrorItem)

const ErrorV2MediaType = "application/vnd.d4l.error.v2+json"

// ErrorResponseV2 writes to the response body an error in the v2 format.
// It also sets the Content-Type to the corresponding media type.
// In case of failure to encode the body, it returns an internal error also in
// format v2. It logs any unexpected error to the given logger.
func ErrorResponseV2(l logger, w http.ResponseWriter, req *http.Request, errorCode ErrorCode, errorMessage string, statusCode int,
	modifiers ...ResponseErrModifier) {
	if err := ErrorV2(w, req, errorCode, errorMessage, statusCode, modifiers...); err != nil {
		_ = l.ErrInternal(req.Context(), err)
	}
}

// WithDetails allows to add some details to the error response object.
func WithDetails(details interface{}) ResponseErrModifier {
	return func(e *ErrorItem) {
		e.Details = details
	}
}

func ErrorV2(w http.ResponseWriter, req *http.Request, errorCode ErrorCode, errorMessage string, statusCode int,
	modifiers ...ResponseErrModifier) error {
	w.Header().Set("Content-Type", ErrorV2MediaType)
	traceID, _ := req.Context().Value(log.TraceIDContextKey).(string)
	errItem := ErrorItem{
		Code:    errorCode,
		TraceID: traceID,
		Message: errorMessage,
	}

	// apply all modifiers
	for _, m := range modifiers {
		m(&errItem)
	}

	body, err := json.Marshal(ErrorResponse{
		Errors: []ErrorItem{errItem},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errBody, err2 := json.Marshal(InternalErrorResponse(traceID))
		if err2 != nil {
			return fmt.Errorf("could not marshal the internal error response: %s: %w", err2, err)
		}
		_, _ = w.Write(errBody)
		return fmt.Errorf("could not marshal the JSON error response: %w", err)
	}
	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
	return nil
}

func InternalErrorResponse(traceID string) ErrorResponse {
	return ErrorResponse{
		Errors: []ErrorItem{
			{
				Code:    InternalError,
				TraceID: traceID,
				Message: ErrorMessages[InternalError],
			},
		},
	}
}
