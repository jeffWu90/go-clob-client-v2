package clob

import (
	"errors"
	"fmt"
)

// ErrL1AuthUnavailable is returned when an L1-only endpoint is invoked but no signer was configured.
var ErrL1AuthUnavailable = errors.New("signer is needed to interact with this endpoint")

// ErrL2AuthUnavailable is returned when an L2-only endpoint is invoked but no API credentials are set.
var ErrL2AuthUnavailable = errors.New("API credentials are needed to interact with this endpoint")

// ApiError is returned by HTTP helpers when the CLOB server replies with a non-2xx status.
// Status is the raw HTTP status code; Data is the parsed JSON body (any shape).
type ApiError struct {
	Message string
	Status  int
	Data    any
}

func (e *ApiError) Error() string {
	if e.Status != 0 {
		return fmt.Sprintf("api error (%d): %s", e.Status, e.Message)
	}
	return fmt.Sprintf("api error: %s", e.Message)
}

// NewApiError constructs an ApiError with the given message, status, and decoded body.
func NewApiError(message string, status int, data any) *ApiError {
	return &ApiError{Message: message, Status: status, Data: data}
}
