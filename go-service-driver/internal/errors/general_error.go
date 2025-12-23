package errors

import "net/http"

func NewInternalServerError(err error) *AppError {
	return &AppError{
		Code:     ErrInternalServer,
		Message:  "Internal server error",
		Status:   http.StatusInternalServerError,
		Internal: err,
	}
}

func NewInvalidRequestError(message string) *AppError {
	return &AppError{
		Code:    ErrInvalidRequest,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

func NewRateLimitExceededError() *AppError {
	return &AppError{
		Code:    ErrRateLimitExceeded,
		Message: "Rate limit exceeded",
		Status:  http.StatusTooManyRequests,
	}
}

func NewDatabaseError(err error) *AppError {
	return &AppError{
		Code:     ErrDatabaseError,
		Message:  "Database error occurred",
		Status:   http.StatusInternalServerError,
		Internal: err,
	}
}

func NewRedisError(err error) *AppError {
	return &AppError{
		Code:     ErrRedisError,
		Message:  "Redis error occurred",
		Status:   http.StatusInternalServerError,
		Internal: err,
	}
}

func NewParseResponseError(err error) *AppError {
	return &AppError{
		Code:     ErrParseResponse,
		Message:  "Failed to parse response",
		Status:   http.StatusInternalServerError,
		Internal: err,
	}
}
