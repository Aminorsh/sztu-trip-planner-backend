package errors

import "net/http"

func NewJSONArrayScanError() *AppError {
	return &AppError{
		Code:    ErrInvalidRequest,
		Message: "Failed to scan JSON array from database",
		Status:  http.StatusBadRequest,
	}
}

func NewAmapConfigError() *AppError {
	return &AppError{
		Code:    ErrAmapConfig,
		Message: "AMAP API configuration is missing or invalid",
		Status:  http.StatusInternalServerError,
	}
}

func NewAmapAPIError(err error) *AppError {
	return &AppError{
		Code:     ErrAmapAPI,
		Message:  "Error occurred while calling AMAP API",
		Status:   http.StatusInternalServerError,
		Internal: err,
	}
}

func NewReadResponseError(err error) *AppError {
	return &AppError{
		Code:     ErrReadResponse,
		Message:  "Failed to read response from AMAP API",
		Status:   http.StatusInternalServerError,
		Internal: err,
	}
}

func NewParseResponseError(err error) *AppError {
	return &AppError{
		Code:     ErrParseResponse,
		Message:  "Failed to parse response from AMAP API",
		Status:   http.StatusInternalServerError,
		Internal: err,
	}
}

func NewPlaceNotFoundError() *AppError {
	return &AppError{
		Code:    ErrPlaceNotFound,
		Message: "Place not found",
		Status:  http.StatusNotFound,
	}
}
