package entity

import "time"

// Comment -.
type Comment struct {
	ID        string    `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	PostID    string    `json:"post_id"    example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    string    `json:"user_id"    example:"550e8400-e29b-41d4-a716-446655440000"`
	Body      string    `json:"body"       example:"Nice post!"`
	CreatedAt time.Time `json:"created_at" example:"2026-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-01-01T00:00:00Z"`
} // @name entity.Comment
