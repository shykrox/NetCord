package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrNotFound           = errors.New("resource not found")
	ErrForbidden          = errors.New("forbidden")
	ErrUnavailable        = errors.New("service unavailable")
	ErrVoiceNotConfigured = errors.New("voice not configured")
)

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}
