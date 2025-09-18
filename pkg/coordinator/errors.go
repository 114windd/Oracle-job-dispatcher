package coordinator

import (
	"encoding/json"
	"net/http"
)

// APIError represents a structured error response
type APIError struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

// ErrorResponse represents the JSON structure for error responses
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

// NewAPIError creates a new APIError with the given parameters
func NewAPIError(message string, code int, details string) *APIError {
	return &APIError{
		Error:   message,
		Code:    code,
		Details: details,
	}
}

// WriteError writes a structured error response to the HTTP response writer
func WriteError(w http.ResponseWriter, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)

	response := ErrorResponse{
		Error:   err.Error,
		Code:    err.Code,
		Details: err.Details,
	}

	json.NewEncoder(w).Encode(response)
}

// WriteJSONError is a convenience function to write error responses
func WriteJSONError(w http.ResponseWriter, message string, code int, details string) {
	err := NewAPIError(message, code, details)
	WriteError(w, err)
}

// Common error types
var (
	ErrRateLimitExceeded = &APIError{
		Error:   "rate limit exceeded",
		Code:    429,
		Details: "too many requests from this IP address",
	}

	ErrWorkerTimeout = &APIError{
		Error:   "worker timeout",
		Code:    504,
		Details: "one or more workers did not respond within the timeout period",
	}

	ErrNoWorkersAvailable = &APIError{
		Error:   "no workers available",
		Code:    503,
		Details: "no workers are currently available to process the request",
	}

	ErrInvalidRequest = &APIError{
		Error:   "invalid request",
		Code:    400,
		Details: "the request format is invalid or missing required fields",
	}

	ErrInternalServer = &APIError{
		Error:   "internal server error",
		Code:    500,
		Details: "an unexpected error occurred while processing the request",
	}

	ErrNATSConnection = &APIError{
		Error:   "message queue unavailable",
		Code:    503,
		Details: "unable to connect to the message queue system",
	}
)
