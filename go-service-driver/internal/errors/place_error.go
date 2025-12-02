package errors

import "net/http"

func NewJSONArrayScanError() *AppError {
	return &AppError{
		Code:    ErrInvalidRequest,
		Message: "Failed to scan JSON array from database",
		Status:  http.StatusBadRequest,
	}
}
