package errors

import "net/http"

func NewUserNotFoundError() *AppError {
	return &AppError{
		Code:    ErrUserNotFound,
		Message: "User not found",
		Status:  http.StatusNotFound,
	}
}

func NewUserAlreadyExistsError(field string) *AppError {
	return &AppError{
		Code:    ErrUserAlreadyExists,
		Message: "User already exists",
		Status:  http.StatusConflict,
		Details: map[string]string{"field": field},
	}
}

func NewInvalidCredentialsError() *AppError {
	return &AppError{
		Code:    ErrInvalidCredentials,
		Message: "Username or password is incorrect",
		Status:  http.StatusUnauthorized,
	}
}

func NewAccountSuspendedError() *AppError {
	return &AppError{
		Code:    ErrAccountSuspended,
		Message: "Account is suspended",
		Status:  http.StatusForbidden,
	}
}

func NewInvalidVerifyCodeError() *AppError {
	return &AppError{
		Code:    ErrInvalidVerifyCode,
		Message: "Invalid verification code",
		Status:  http.StatusBadRequest,
	}
}

func NewVerifyCodeExpiredError() *AppError {
	return &AppError{
		Code:    ErrVerifyCodeExpired,
		Message: "Verification code has expired",
		Status:  http.StatusBadRequest,
	}
}

func NewPasswordTooWeakError() *AppError {
	return &AppError{
		Code:    ErrPasswordTooWeak,
		Message: "Password is too weak",
		Status:  http.StatusBadRequest,
	}
}

func NewIncorrectPasswordError() *AppError {
	return &AppError{
		Code:    ErrIncorrectPassword,
		Message: "Current password is incorrect",
		Status:  http.StatusBadRequest,
	}
}

func NewUnauthorizedError() *AppError {
	return &AppError{
		Code:    ErrUnauthorized,
		Message: "Unauthorized access",
		Status:  http.StatusUnauthorized,
	}
}

func NewGenerateTokenError(err error) *AppError {
	return &AppError{
		Code:     ErrGenerateToken,
		Message:  "Failed to generate authentication token",
		Status:   http.StatusInternalServerError,
		Internal: err,
	}
}
