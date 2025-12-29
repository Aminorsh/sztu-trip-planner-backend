package errors

import "net/http"

func NewTripNotFoundError() *AppError {
	return &AppError{
		Code:    ErrTripNotFound,
		Message: "Trip not found",
		Status:  http.StatusNotFound,
	}
}

func NewTripAccessDeniedError() *AppError {
	return &AppError{
		Code:    ErrTripAccessDenied,
		Message: "Access to trip denied",
		Status:  http.StatusForbidden,
	}
}

func NewInvalidTripDataError() *AppError {
	return &AppError{
		Code:    ErrInvalidTripData,
		Message: "Invalid trip data provided",
		Status:  http.StatusBadRequest,
	}
}

func NewTripItemNotFoundError() *AppError {
	return &AppError{
		Code:    ErrTripItemNotFound,
		Message: "Trip item not found",
		Status:  http.StatusNotFound,
	}
}
