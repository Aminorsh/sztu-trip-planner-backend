package errors

import (
	"fmt"
)

type AppError struct {
	Code     string            `json:"code"`
	Message  string            `json:"message"`
	Status   int               `json:"-"`
	Details  map[string]string `json:"details,omitempty"`
	Internal error             `json:"-"`
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

const (
	// General Errors
	ErrInternalServer    = "INTERNAL_SERVER_ERROR"
	ErrInvalidRequest    = "INVALID_REQUEST"
	ErrRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrDatabaseError     = "DATABASE_ERROR"
	ErrRedisError        = "REDIS_ERROR"

	// User Errors
	ErrUserNotFound       = "USER_NOT_FOUND"
	ErrUserAlreadyExists  = "USER_ALREADY_EXISTS"
	ErrInvalidCredentials = "INVALID_CREDENTIALS"
	ErrAccountSuspended   = "ACCOUNT_SUSPENDED"
	ErrInvalidVerifyCode  = "INVALID_VERIFY_CODE"
	ErrVerifyCodeExpired  = "VERIFY_CODE_EXPIRED"
	ErrPasswordTooWeak    = "PASSWORD_TOO_WEAK"
	ErrIncorrectPassword  = "INCORRECT_PASSWORD"
	ErrUnauthorized       = "UNAUTHORIZED"
	ErrGenerateToken      = "TOKEN_GENERATION_FAILED"

	// Trip Errors
	ErrTripNotFound     = "TRIP_NOT_FOUND"
	ErrTripAccessDenied = "TRIP_ACCESS_DENIED"
	ErrInvalidTripData  = "INVALID_TRIP_DATA"
	ErrTripItemNotFound = "TRIP_ITEM_NOT_FOUND"

	// Place Errors
	ErrJSONArrayScan = "JSON_ARRAY_SCAN_ERROR"
	ErrAmapConfig    = "AMAP_CONFIG_ERROR"
	ErrAmapAPI       = "AMAP_API_ERROR"
	ErrReadResponse  = "READ_RESPONSE_ERROR"
	ErrParseResponse = "PARSE_RESPONSE_ERROR"
	ErrPlaceNotFound = "PLACE_NOT_FOUND"
)
