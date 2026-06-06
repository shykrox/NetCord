package service

import "errors"

var ErrInvalidCredentials = errors.New("invalid email or password")

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}
