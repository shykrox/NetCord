package repository

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserConflict           = errors.New("username or email already exists")
	ErrServerNotFound         = errors.New("server not found")
	ErrChannelNotFound        = errors.New("channel not found")
	ErrMessageNotFound        = errors.New("message not found")
	ErrForbidden              = errors.New("forbidden")
	ErrChannelConflict        = errors.New("channel already exists")
	ErrAttachmentNotFound     = errors.New("attachment not found")
	ErrFriendRequestConflict  = errors.New("friend request already exists")
	ErrFriendRequestNotFound  = errors.New("friend request not found")
	ErrFriendNotFound         = errors.New("friend not found")
	ErrDMConversationNotFound = errors.New("dm conversation not found")
	ErrRoleNotFound           = errors.New("role not found")
	ErrRoleConflict           = errors.New("role already exists")
	ErrInviteNotFound         = errors.New("invite not found")
	ErrInviteUnavailable      = errors.New("invite unavailable")
	ErrAIJobNotFound          = errors.New("ai job not found")
)
