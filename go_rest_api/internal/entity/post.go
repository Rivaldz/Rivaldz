package entity

import "time"

// Post -.
type Post struct {
	ID           string    `json:"id"            example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID       string    `json:"user_id"       example:"550e8400-e29b-41d4-a716-446655440000"`
	Title        string    `json:"title"         example:"My first post"`
	Body         string    `json:"body"          example:"Post content"`
	CommentCount int       `json:"comment_count" example:"3"`
	CreatedAt    time.Time `json:"created_at"    example:"2026-01-01T00:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at"    example:"2026-01-01T00:00:00Z"`
} // @name entity.Post
