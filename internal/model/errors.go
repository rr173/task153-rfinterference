package model

import "fmt"

const (
	CodeValidation        = "VALIDATION_FAILED"
	CodeUnknownStation    = "UNKNOWN_STATION"
	CodeDisabledStation   = "STATION_DISABLED"
	CodeConflict          = "IDEMPOTENCY_CONFLICT"
	CodeNotFound          = "NOT_FOUND"
	CodeArchived          = "EVENT_ARCHIVED"
	CodeFutureObservation = "FUTURE_OBSERVATION"
	CodeInternal          = "INTERNAL"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

func NewError(code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

func CodeOf(err error) string {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return CodeInternal
}
