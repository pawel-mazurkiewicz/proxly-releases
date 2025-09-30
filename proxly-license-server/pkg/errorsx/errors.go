package errorsx

import (
	"fmt"
	"net/http"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

func NewError(code int, message string, details ...string) *Error {
	err := &Error{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

var (
	ErrInvalidLicenseKey    = NewError(http.StatusBadRequest, "Invalid license key")
	ErrLicenseNotFound      = NewError(http.StatusNotFound, "License not found")
	ErrLicenseRevoked       = NewError(http.StatusForbidden, "License has been revoked")
	ErrLicenseExpired       = NewError(http.StatusForbidden, "License has expired")
	ErrLicenseSuspended     = NewError(http.StatusForbidden, "License is suspended")
	ErrActivationExceeded   = NewError(http.StatusTooManyRequests, "License activation limit exceeded")
	ErrInvalidSignature     = NewError(http.StatusUnauthorized, "Invalid request signature")
	ErrTimestampSkew        = NewError(http.StatusUnauthorized, "Request timestamp too skewed")
	ErrRateLimitExceeded    = NewError(http.StatusTooManyRequests, "Rate limit exceeded")
	ErrInternalServer       = NewError(http.StatusInternalServerError, "Internal server error")
	ErrDatabaseConnection   = NewError(http.StatusServiceUnavailable, "Database connection error")
	ErrInvalidPayload       = NewError(http.StatusBadRequest, "Invalid request payload")
	ErrUnauthorized         = NewError(http.StatusUnauthorized, "Unauthorized")
	ErrForbidden            = NewError(http.StatusForbidden, "Forbidden")
)