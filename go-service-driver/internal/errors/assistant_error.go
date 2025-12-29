package errors

import (
	"net/http"
)

func NewAssistantConfigError() *AppError {
	return &AppError{
		Code:    ErrAssistantConfig,
		Message: "Assistant API configuration is missing or invalid",
		Status:  http.StatusInternalServerError,
	}
}

func NewSerializationError(err error) *AppError {
	return &AppError{
		Code:     ErrSerialization,
		Message:  "Failed to serialize request body for Assistant API",
		Status:   http.StatusBadRequest,
		Internal: err,
	}
}

func NewDeepseekAPIError(err error) *AppError {
	return &AppError{
		Code:     ErrDeepseekAPI,
		Message:  "Failed to call Deepseek API",
		Status:   http.StatusBadRequest,
		Internal: err,
	}
}

func MapDeepseekErrorResponse(statusCode int) *AppError {
	switch statusCode {
	case http.StatusBadRequest:
		return &AppError{
			Code:    ErrDeepseekAPI,
			Message: "Invalid request body format for Deepseek API",
			Status:  http.StatusBadRequest,
		}
	case http.StatusUnauthorized:
		return &AppError{
			Code:    ErrAuthenticationFails,
			Message: "Authentication fails due to the wrong API key",
			Status:  http.StatusUnauthorized,
		}
	case http.StatusPaymentRequired:
		return &AppError{
			Code:    ErrInsufficientBalance,
			Message: "Insufficient balance for Deepseek API",
			Status:  http.StatusPaymentRequired,
		}
	case http.StatusUnprocessableEntity:
		return &AppError{
			Code:    ErrInvalidParameters,
			Message: "Invalid parameters in Deepseek API request",
			Status:  http.StatusUnprocessableEntity,
		}
	case http.StatusTooManyRequests:
		return &AppError{
			Code:    ErrRateLimitExceeded,
			Message: "Rate limit reached for Deepseek API",
			Status:  http.StatusTooManyRequests,
		}
	case http.StatusInternalServerError:
		return &AppError{
			Code:    ErrInternalServer,
			Message: "Deepseek API server error",
			Status:  http.StatusInternalServerError,
		}
	case http.StatusServiceUnavailable:
		return &AppError{
			Code:    ErrServerOverloaded,
			Message: "Deepseek API server overloaded",
			Status:  http.StatusServiceUnavailable,
		}
	default:
		return &AppError{
			Code:    ErrDeepseekAPI,
			Message: "Unexpected error from Deepseek API",
			Status:  http.StatusBadRequest,
		}
	}
}
