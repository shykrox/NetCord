package repository

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserConflict       = errors.New("username or email already exists")
	ErrServerNotFound     = errors.New("server not found")
	ErrChannelNotFound    = errors.New("channel not found")
	ErrChannelConflict    = errors.New("channel already exists")
	ErrAttachmentNotFound = errors.New("attachment not found")
)
