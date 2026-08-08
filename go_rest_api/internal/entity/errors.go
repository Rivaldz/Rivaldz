package entity

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTaskNotFound       = errors.New("task not found")
	ErrTaskForbidden      = errors.New("task does not belong to user")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrPostNotFound       = errors.New("post not found")
	ErrPostForbidden      = errors.New("post does not belong to user")
	ErrCommentNotFound    = errors.New("comment not found")
	ErrCommentForbidden   = errors.New("comment does not belong to user")
)
