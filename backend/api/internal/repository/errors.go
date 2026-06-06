package repository

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserConflict = errors.New("username or email already exists")
)
