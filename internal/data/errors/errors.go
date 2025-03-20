package errors

import (
	"errors"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
)

var (
	// Common OAuth errors
	ErrInvalidGrant        = errors.New("invalid grant")
	ErrMissingRefreshToken = errors.New("missing refresh token")
	ErrInvalidToken        = errors.New("invalid token")
	ErrConfigNotFound      = errors.New("oauth config not found")
	ErrAuthCodeInvalid     = errors.New("authorization code invalid")
	ErrUserInfoUnavailable = errors.New("user info unavailable")
	ErrInvalidCredential   = errors.New("invalid credential")
)

// MapToExternalError maps internal errors to API-friendly external errors
func MapToExternalError(err error) error {
	if errors.Is(err, ErrConfigNotFound) {
		return iam_v1.ErrorServiceFailed("OAuth configuration unavailable")
	}
	if errors.Is(err, ErrAuthCodeInvalid) {
		return iam_v1.ErrorInvalidRequest("Invalid authorization code")
	}
	if errors.Is(err, ErrUserInfoUnavailable) {
		return iam_v1.ErrorServiceFailed("Unable to retrieve user information")
	}
	if errors.Is(err, ErrInvalidToken) {
		return iam_v1.ErrorServiceFailed("Invalid token")
	}
	if errors.Is(err, ErrMissingRefreshToken) {
		return iam_v1.ErrorNeedReauthorization("Refresh token missing, re-authorization required")
	}
	if errors.Is(err, ErrInvalidGrant) {
		return iam_v1.ErrorNeedReauthorization("Invalid or expired grant, re-authorization required")
	}
	if errors.Is(err, ErrInvalidCredential) {
		return iam_v1.ErrorInvalidRequest("invalid credential")
	}

	// Default error
	return iam_v1.ErrorServiceFailed("service error: %v", err.Error())
}
